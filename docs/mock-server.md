# Mock SuccessFactors server

This page is for developers of `baton-successfactors`. It is not part of the public connector docs (`connector.mdx`).

## 1. What the mock is and is not

The mock is a small HTTP server that answers the two SuccessFactors endpoints the connector calls:

- `POST /oauth/token` (SAML bearer grant)
- `GET /odata/v2/EmpJob` (users, in pages)

It lets you run the real connector binary and the `baton` CLI end to end, with no tenant and no secrets. CI uses it in the `mock-test` job.

The mock is not a replacement for a test against a real tenant:

- The 400 body for "SAML assertion in the URL" and the 401 body for "unknown company" are copied from the real service.
- All other error texts are synthetic. We do not know how the real service words them. The code says so.
- The mock does not parse the full OData grammar. It checks only that the first page asks for `$format=json`, `$select`, and `$expand` with `userNav` and `employmentNav`.
- By default the mock accepts a signature made with an expired certificate, because the real behavior is unknown. Use `-reject-expired-cert` to test the other case.

Connector errors include the error message that SuccessFactors sends (the token endpoint and the OData endpoint use two different JSON shapes, and the mock uses both).

Code: `internal/mocksf` (handler and fixtures), `test/mock-successfactors` (command), `test/verify-mock.sh` (the check that CI runs).

## 2. Prerequisites

- Go (the version in `go.mod`)
- `openssl`, `jq`, `curl`
- A `baton` CLI **with the same version as `baton-sdk` in `go.mod`**. An older CLI cannot read the `sync.c1z` file. Run `./scripts/get-baton-sdk.sh` in a scratch directory to download the matching version, and put the `baton` file on your `PATH`.

## 3. Quick start

Terminal 1, in the repo root:

```bash
make mock
```

This builds the mock, makes a throwaway certificate and key in `.mock/` (not committed), prints the `BATON_*` exports, and starts the mock on `127.0.0.1:8089`.

Terminal 2: paste the printed `export` lines, then:

```bash
go run ./cmd/baton-successfactors
baton resources
```

Expected result: 12 users, including `rsmith`.

To run all the checks that CI runs, use `BATON=$(command -v baton) make test-mock`.

## 4. Step by step (no make)

```bash
mkdir -p .mock
openssl req -x509 -newkey rsa:2048 -nodes -days 1 -subj /CN=mock \
  -keyout .mock/key.pem -out .mock/cert.pem
go run ./test/mock-successfactors -cert .mock/cert.pem
```

In a second terminal, set the variables. The values are the defaults of the mock. `BATON_PUBLIC_KEY` and `BATON_PRIVATE_KEY` are **file paths**, not PEM contents:

```bash
export BATON_INSTANCE_URL=http://127.0.0.1:8089
export BATON_COMPANY_ID=MOCKCO
export BATON_CID=mock-client
export BATON_SAML_API_KEY=mock-api-key
export BATON_ISSUER_URL=https://mock-issuer.example
export BATON_SUBJECT_NAME_ID=100001
export BATON_PUBLIC_KEY="$PWD/.mock/cert.pem"
export BATON_PRIVATE_KEY="$PWD/.mock/key.pem"
go run ./cmd/baton-successfactors
baton resources
```

The mock flags `-company`, `-client-id`, `-api-key`, `-issuer`, and `-subject` change the values that the mock expects. Change the matching `BATON_*` variable too.

## 5. Run against a C1 dev instance (self-hosted mode)

This section needs credentials for a C1 dev tenant. CI does not run it.

1. In the dev tenant, create a SuccessFactors connector in self-hosted mode. Create the client ID and secret as described in the self-hosted section of `connector.mdx`.
2. Start the mock (section 3 or 4).
3. Run the connector on your machine with the mock exports and the dev tenant credentials:

   ```bash
   export BATON_CLIENT_ID=<client id from the dev tenant>
   export BATON_CLIENT_SECRET=<client secret from the dev tenant>
   go run ./cmd/baton-successfactors
   ```

   The connector runs as a service, syncs from the mock, and uploads the result.
4. In the dev tenant, check that the app shows the 12 fixture users with the expected statuses (`bwong` and `cmuller` are disabled).

Limits:

- This works only in self-hosted mode on your machine. A connector that C1 hosts runs in the C1 cloud and cannot reach a mock on `localhost`.
- Use a dev tenant only. Never use real customer data in the fixtures.

## 6. Run the connector in Docker against the mock

Use an image that includes the key-file change (the `--public-key` and `--private-key` values are file paths). Replace `<tag>` with a release that includes it.

Start the mock so that it listens on all interfaces and builds `__next` links with the host name that the container uses:

```bash
go run ./test/mock-successfactors -cert .mock/cert.pem \
  -addr 0.0.0.0:8089 -public-url http://host.docker.internal:8089
```

Run the connector image:

```bash
docker run --rm \
  --add-host=host.docker.internal:host-gateway \
  -v "$PWD/.mock:/keys" -v "$PWD:/out" \
  -e BATON_INSTANCE_URL=http://host.docker.internal:8089 \
  -e BATON_COMPANY_ID=MOCKCO -e BATON_CID=mock-client \
  -e BATON_SAML_API_KEY=mock-api-key \
  -e BATON_ISSUER_URL=https://mock-issuer.example \
  -e BATON_SUBJECT_NAME_ID=100001 \
  -e BATON_PUBLIC_KEY=/keys/cert.pem -e BATON_PRIVATE_KEY=/keys/key.pem \
  public.ecr.aws/conductorone/baton-successfactors:<tag> -f /out/sync.c1z
baton -f sync.c1z resources
```

`--add-host` is needed on Linux. Docker Desktop on macOS and Windows already knows `host.docker.internal`.

## 7. Test scenarios

Each example starts the mock with one change. The connector uses the exports from section 4.

| Case | Mock command | Result |
|---|---|---|
| Expired certificate, accepted (default) | `openssl req -x509 -newkey rsa:2048 -nodes -days 1 -subj /CN=mock -keyout .mock/key.pem -out .mock/cert.pem -not_after 20200101000000Z`, then `go run ./test/mock-successfactors -cert .mock/cert.pem` | The sync passes |
| Expired certificate, rejected | The same certificate, with `-reject-expired-cert` added | The sync fails with 401 |
| Wrong company ID | Run the connector with `BATON_COMPANY_ID=OTHER` | The sync fails with 401, and the connector error shows "Company OTHER does not exist in this data center" |
| Old request style (assertion in the URL) | `curl -s -X POST "http://127.0.0.1:8089/oauth/token?assertion=x"` | 400, "SAML assertion is not allowed in URL parameter." |
| Bad date | `go run ./test/mock-successfactors -cert .mock/cert.pem -scenario bad-date` | The sync fails with "failed to build user resource" |
| 404 on a page | `go run ./test/mock-successfactors -cert .mock/cert.pem -fail-page 2 -fail-status 404` | The sync fails with "user list returned not found"; C1 keeps the previous data |
| Error on page 2 | `go run ./test/mock-successfactors -cert .mock/cert.pem -fail-page 2 -fail-status 400` | The sync fails on the second page |
| Custom data | `go run ./test/mock-successfactors -cert .mock/cert.pem -fixtures ./my-users.json` | The sync returns your users |
| No users | `go run ./test/mock-successfactors -cert .mock/cert.pem -scenario empty` | The sync passes with 0 users |

Other flags: `-page-size` (default 5), `-addr`, `-public-url`.

Choose the status for `-fail-status` with care. The connector client retries a 5xx status, such as 500, for many minutes, so the sync seems to hang (a run did not end in 7 minutes). A 404 on any page fails the sync with "user list returned not found", so C1 keeps the previous data. Use 400 or 404 to get a quick, clear failure.

For the record, the baton-sdk syncer retries a 5xx status and 429 without a limit (waits of 1s, 2s, 3s, and so on, capped at 60s) and logs "retrying operation". Use `-fail-status 503` to see those logs, and stop the connector with Ctrl-C. Do not use a 5xx status in an automated check.

The `-not_after` option of `openssl req` needs OpenSSL 3.4 or newer. With an older version, use `faketime` or sign with `openssl ca`.

## 8. Custom fixtures

`-fixtures` takes a JSON array of EmpJob items. Use `internal/mocksf/testdata/users.json` as the template. The field names are the JSON names in `pkg/connector/client/models.go`.

- Dates use the SuccessFactors format `/Date(<milliseconds since 1970>)/`. An offset such as `/Date(1735603200000+0000)/` is also accepted.
- The status comes from `employmentNav.endDate`. No end date, or an end date in the future, gives `enabled`. An end date in the past gives `disabled`. A value without digits fails the sync.
- Linked objects (`managerUserNav`, `companyNav`, `positionNav`, and so on) can be `null` or missing. The profile fields are then empty strings.
- `emplStatusNav` and `employeeClassNav` give the first label of `picklistLabels.results`, or an empty string if there are none.
- Use synthetic data only. The default fixtures use the BestRun demo style and the reserved domain `bestrunsap.example`.

## 9. Troubleshooting

| Symptom | Cause |
|---|---|
| `cannot access file` | `BATON_PUBLIC_KEY` or `BATON_PRIVATE_KEY` holds PEM contents. Use a file path. |
| 401 and the mock log shows `POST /oauth/token -> 401` for a signed request | The mock `-cert` is not the certificate in `BATON_PUBLIC_KEY`, or the issuer, subject, API key, or client ID differs from the mock flags. |
| 401, "Company ... does not exist" | `BATON_COMPANY_ID` does not match the mock `-company`. |
| `connection refused` | The mock is not running, or (in Docker) the host name does not reach your machine. See section 6. |
| `address already in use` | Another process uses the port. Use `-addr` and change `BATON_INSTANCE_URL`. |
| `c1z: invalid file` from `baton` | The `baton` CLI is older than the `baton-sdk` version in `go.mod`. Run `./scripts/get-baton-sdk.sh`. |
