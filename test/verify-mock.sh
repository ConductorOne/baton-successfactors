#!/usr/bin/env bash
# Run baton-successfactors against the mock SuccessFactors server and check the result.
#
# Environment:
#   BATON_SUCCESSFACTORS  connector binary (absolute path)
#   MOCK                  mock server binary (absolute path)
#   BATON                 baton CLI, at or above the baton-sdk version in go.mod (default: baton)
#   WORKDIR               work directory (default: .mock)
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
: "${BATON_SUCCESSFACTORS:?set BATON_SUCCESSFACTORS to the connector binary}"
: "${MOCK:?set MOCK to the mock server binary}"
BATON="${BATON:-baton}"
WORKDIR="${WORKDIR:-.mock}"
mkdir -p "$WORKDIR"
WORKDIR="$(cd "$WORKDIR" && pwd)"
PORT=8089
BAD_PORT=8090
NOTFOUND_PORT=8091
NOTFOUND_FIRST_PORT=8092
URL="http://127.0.0.1:${PORT}"
EXPECTED="$root/internal/mocksf/testdata/expected.json"

command -v openssl >/dev/null || { echo "openssl is required" >&2; exit 1; }
command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }
command -v curl >/dev/null || { echo "curl is required" >&2; exit 1; }
# Guard each sync: the baton-sdk syncer retries a 5xx or 429 without limit.
# macOS has no timeout by default, so the guard is optional.
if command -v timeout >/dev/null; then guard=(timeout 120); else guard=(); fi
command -v "$BATON" >/dev/null || { echo "baton CLI not found (set BATON)" >&2; exit 1; }

# A throwaway cert and key. The mock trusts this cert.
gen_cert() { # dir
  mkdir -p "$1"
  openssl req -x509 -newkey rsa:2048 -nodes -days 1 -subj /CN=mock \
    -keyout "$1/key.pem" -out "$1/cert.pem" 2>/dev/null
}
gen_cert "$WORKDIR"
gen_cert "$WORKDIR/other"

pids=()
cleanup() {
  local rc=$?
  if [ "$rc" -ne 0 ]; then
    echo "--- mock log ---" >&2
    cat "$WORKDIR"/mock*.log >&2 2>/dev/null || true
  fi
  if [ ${#pids[@]} -gt 0 ]; then kill "${pids[@]}" 2>/dev/null || true; fi
}
trap cleanup EXIT

start_mock() { # port log [extra flags...]
  local port="$1" log="$2"
  shift 2
  "$MOCK" -addr "127.0.0.1:${port}" -cert "$WORKDIR/cert.pem" "$@" >"$log" 2>&1 &
  local pid=$!
  pids+=("$pid")
  for _ in $(seq 1 50); do
    # If the port is already in use, our mock exits and the old server would answer.
    if ! kill -0 "$pid" 2>/dev/null; then
      echo "mock exited on port ${port} (is the port in use? see $log)" >&2
      cat "$log" >&2
      return 1
    fi
    curl -fs "http://127.0.0.1:${port}/healthz" >/dev/null 2>&1 && return 0
    sleep 0.2
  done
  echo "mock did not start on port ${port}" >&2
  return 1
}

eval "$("$root/test/mock-env.sh" "$WORKDIR" "$URL")"

start_mock "$PORT" "$WORKDIR/mock.log"

# run_connector <dir> [VAR=value ...]: run the connector in a fresh dir, output in <dir>/out.log.
run_connector() {
  local dir="$1"
  shift
  rm -rf "$dir"
  mkdir -p "$dir"
  (cd "$dir" && env "$@" BATON_LOG_LEVEL=info ${guard[@]+"${guard[@]}"} "$BATON_SUCCESSFACTORS") >"$dir/out.log" 2>&1
}

# expect_fail <name> <pattern> <dir> [VAR=value ...]: the sync must fail and the output must match.
expect_fail() {
  local name="$1" pattern="$2" dir="$3"
  shift 3
  if run_connector "$dir" "$@"; then
    echo "FAIL: ${name}: the sync passed, but it must fail" >&2
    exit 1
  fi
  if ! grep -q -- "$pattern" "$dir/out.log"; then
    echo "FAIL: ${name}: the output has no match for: ${pattern}" >&2
    tail -n 5 "$dir/out.log" >&2
    exit 1
  fi
  echo "ok: ${name}"
}

echo "== positive: sync, verify-resources.sh, expected.json"
pos="$WORKDIR/run-positive"
rm -rf "$pos"
mkdir -p "$pos"
(cd "$pos" && BATON="$BATON" BATON_PRINCIPAL=rsmith ${guard[@]+"${guard[@]}"} "$root/test/verify-resources.sh") >"$pos/out.log" 2>&1 \
  || { tail -n 20 "$pos/out.log" >&2; echo "FAIL: verify-resources.sh" >&2; exit 1; }
echo "ok: verify-resources.sh"

(cd "$pos" && "$BATON" resources --output-format=json) >"$pos/resources.json"
mismatches="$(jq -r --slurpfile exp "$EXPECTED" '
  ($exp[0]) as $e
  | ([.resources[].resource]
      | map({key: .displayName,
             value: {status: (.status.status | sub("RESOURCE_STATUS_"; "") | ascii_downcase),
                     profile: .profile}})
      | from_entries) as $got
  | ( (if ($got | length) != $e.count then ["count: got \($got | length), want \($e.count)"] else [] end)
    + [ $e.users | to_entries[] | . as $u
        | ($got[$u.key]) as $g
        | if $g == null then "missing user \($u.key)"
          elif $g.status != $u.value.status then "status \($u.key): got \($g.status), want \($u.value.status)"
          else ($u.value.profile | to_entries[]
                | select($g.profile[.key] != .value)
                | "profile \($u.key).\(.key): got \($g.profile[.key] | tojson), want \(.value | tojson)")
          end ] )
  | .[]' "$pos/resources.json")"
if [ -n "$mismatches" ]; then
  echo "FAIL: the synced users differ from expected.json:" >&2
  echo "$mismatches" >&2
  exit 1
fi
echo "ok: $(jq '.resources | length' "$pos/resources.json") users match expected.json (key files)"

# Positive run B (backward compatibility): the keys as PEM contents in BATON_PUBLIC_KEY and
# BATON_PRIVATE_KEY, as in earlier releases.
run_connector "$WORKDIR/run-contents" -u BATON_PUBLIC_KEY_FILE -u BATON_PRIVATE_KEY_FILE \
  "BATON_PUBLIC_KEY=$(cat "$WORKDIR/cert.pem")" "BATON_PRIVATE_KEY=$(cat "$WORKDIR/key.pem")" \
  || { tail -n 20 "$WORKDIR/run-contents/out.log" >&2; echo "FAIL: sync with PEM contents" >&2; exit 1; }
contents_users="$(cd "$WORKDIR/run-contents" && "$BATON" resources --output-format=json | jq '.resources | length')"
want_users="$(jq '.count' "$EXPECTED")"
[ "$contents_users" = "$want_users" ] \
  || { echo "FAIL: PEM contents: got $contents_users users, want $want_users" >&2; exit 1; }
echo "ok: ${contents_users} users with the keys as PEM contents"

echo "== negative checks"
# 1. The old request style (assertion in the URL) gets the real 400 body.
code="$(curl -s -o "$WORKDIR/old-style.json" -w '%{http_code}' -X POST \
  "${URL}/oauth/token?assertion=x&company_id=MOCKCO")"
[ "$code" = "400" ] || { echo "FAIL: old-style request: status $code, want 400" >&2; exit 1; }
jq -e '.errorMessage == "SAML assertion is not allowed in URL parameter."' "$WORKDIR/old-style.json" >/dev/null \
  || { echo "FAIL: old-style request: unexpected body" >&2; exit 1; }
echo "ok: assertion in the URL gets 400"

# 2. A wrong company ID makes the mock answer 401, and the sync fails.
before="$(grep -c 'POST /oauth/token -> 401' "$WORKDIR/mock.log" || true)"
expect_fail "wrong company" "does not exist in this data center" "$WORKDIR/run-company" BATON_COMPANY_ID=WRONGCO
after="$(grep -c 'POST /oauth/token -> 401' "$WORKDIR/mock.log" || true)"
[ "$after" -gt "$before" ] || { echo "FAIL: wrong company: the mock did not answer 401" >&2; exit 1; }

# 3. A cert that the mock does not trust gives a signature failure (401).
before="$after"
expect_fail "untrusted cert" "401" "$WORKDIR/run-cert" \
  "BATON_PUBLIC_KEY_FILE=$WORKDIR/other/cert.pem" "BATON_PRIVATE_KEY_FILE=$WORKDIR/other/key.pem"
after="$(grep -c 'POST /oauth/token -> 401' "$WORKDIR/mock.log" || true)"
[ "$after" -gt "$before" ] || { echo "FAIL: untrusted cert: the mock did not answer 401" >&2; exit 1; }

# 4. PEM contents in a *_FILE variable (a path is expected) fail.
expect_fail "PEM contents in the file field" "cannot access file" "$WORKDIR/run-pem" \
  "BATON_PUBLIC_KEY_FILE=$(cat "$WORKDIR/cert.pem")"

# 4b. Both fields of a pair set fail (mutually exclusive).
expect_fail "certificate: text and file both set" "none of the others can be" "$WORKDIR/run-both" \
  "BATON_PUBLIC_KEY=$(cat "$WORKDIR/cert.pem")"
expect_fail "private key: text and file both set" "none of the others can be" "$WORKDIR/run-both-key" \
  "BATON_PRIVATE_KEY=$(cat "$WORKDIR/key.pem")"

# 4c. Neither field of a pair set fails (at least one).
expect_fail "certificate: neither set" "at least one of the flags" "$WORKDIR/run-neither" -u BATON_PUBLIC_KEY_FILE
expect_fail "private key: neither set" "at least one of the flags" "$WORKDIR/run-neither-key" -u BATON_PRIVATE_KEY_FILE

# 5. A user with an unparseable endDate fails the sync.
start_mock "$BAD_PORT" "$WORKDIR/mock-bad-date.log" -scenario bad-date
expect_fail "bad date" "failed to build user resource" "$WORKDIR/run-bad-date" \
  "BATON_INSTANCE_URL=http://127.0.0.1:${BAD_PORT}"

# 6. A 404 from the user list must fail the sync, on a later page and on the first page.
# Without this, the SDK treats NotFound as a warning and the sync ends with a partial list.
start_mock "$NOTFOUND_PORT" "$WORKDIR/mock-404-page2.log" -fail-page 2 -fail-status 404
expect_fail "404 on page 2" "user list returned not found" "$WORKDIR/run-404-page2" \
  "BATON_INSTANCE_URL=http://127.0.0.1:${NOTFOUND_PORT}"
start_mock "$NOTFOUND_FIRST_PORT" "$WORKDIR/mock-404-page1.log" -fail-page 1 -fail-status 404
expect_fail "404 on page 1" "user list returned not found" "$WORKDIR/run-404-page1" \
  "BATON_INSTANCE_URL=http://127.0.0.1:${NOTFOUND_FIRST_PORT}"

echo "All mock checks passed."
