#!/usr/bin/env bash
# Print the BATON_* exports that match the defaults of test/mock-successfactors.
# Usage: eval "$(test/mock-env.sh [workdir] [instance-url])"
set -euo pipefail

dir="$(cd "${1:-.mock}" && pwd)"
url="${2:-http://127.0.0.1:8089}"

cat <<ENV
export BATON_INSTANCE_URL='${url}'
export BATON_COMPANY_ID='MOCKCO'
export BATON_CID='mock-client'
export BATON_SAML_API_KEY='mock-api-key'
export BATON_ISSUER_URL='https://mock-issuer.example'
export BATON_SUBJECT_NAME_ID='100001'
export BATON_PUBLIC_KEY_FILE='${dir}/cert.pem'
export BATON_PRIVATE_KEY_FILE='${dir}/key.pem'
ENV
