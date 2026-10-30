![Baton Logo](./baton-logo.png)

# `baton-successfactors` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-successfactors.svg)](https://pkg.go.dev/github.com/conductorone/baton-successfactors) ![ci](https://github.com/conductorone/baton-successfactors/actions/workflows/ci.yaml/badge.svg) ![verify](https://github.com/conductorone/baton-successfactors/actions/workflows/verify.yaml/badge.svg)

`baton-successfactors` is a connector for SAP SuccessFactors built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It communicates with the OData V2 protocol to sync data about users.

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.
# Credentials
Please reference the [SAP SuccessFactors API Reference Guide- OData V2](https://help.sap.com/doc/a7c08a422cc14e1eaaffee83610a981d/2411/en-US/SF_HCM_OData_API_DEV.pdf) to configure the API credentials on SuccessFactors. To access SuccessFactor, you must provide the following:
1. SAML API key from when you register the connector as an OAuth application in SuccessFactors.
2. Issuer URL from when you register the connector as an OAuth application in SuccessFactors.
3. Client ID of the application from when you register the connector as an OAuth application in SuccessFactors.
4. X.509 Certificate for signing the SAML assertion from when you register the connector as an OAuth application in SuccessFactors. Also provide the matching private key. The `--public-key` and `--private-key` flags (and the `BATON_PUBLIC_KEY` and `BATON_PRIVATE_KEY` environment variables) take the path to a PEM file, not the PEM contents.
5. User ID of the admin user that will be added to the SAML subject. Check out [Person/User IDs Used Within Employee Central - Employee Profile](https://userapps.support.sap.com/sap/support/knowledge/en/2493579) to learn more.
6. Company ID for your company in SuccessFactors tenant. Follow [How to find the SAP SuccessFactors CompanyID](https://userapps.support.sap.com/sap/support/knowledge/en/2655655).
7. Instance URL for SuccessFactor API calls, you can find your API server following SAP documentation here [List of SAP SuccessFactors API server](https://help.sap.com/docs/SAP_SUCCESSFACTORS_PLATFORM/d599f15995d348a1b45ba5603e2aba9b/af2b8d5437494b12be88fe374eba75b6.html)

# Steps to Create Credentials in SuccessFactor

## Create OAuth2 Client Applications
Navigate to Admin tools > Manage OAuth2 Client Applications and select Register Client Application:
![image](https://github.com/user-attachments/assets/595a269e-5c0c-45dc-b16b-85b2d955db63)
Note down your **Company ID** and add:
1. Application Name (e.g. Baton Connector)
2. Application URL (Note down as **Isser URL**, e.g. https://www.conductorone.com/)
3. **X.509 Certificate**, you can choose to generate a new certificate on the platform or locally. If you are uploading a self generated certificate, remove the lines -----BEGIN CERTIFICATE----- and -----END CERTIFICATE-----.

## Retrieving API Key
After creating the OAuth Client Application, select edit application and note down the API Key. Use this API Key as both the **Client ID** and the **SAML API key**.
![image](https://github.com/user-attachments/assets/287c8fb4-4e25-44df-b464-57569adf1473)


# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-successfactors
baton-successfactors --public-key ./sf-cert.pem --private-key ./sf-key.pem
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_COMPANY_ID=company_id -e BATON_CID=client_id -e BATON_SAML_API_KEY=saml_api_key -e BATON_INSTANCE_URL=instance_url -e BATON_ISSUER_URL=issuer_url -e BATON_SUBJECT_NAME_ID=subject_name_id -e BATON_PUBLIC_KEY=/out/sf-cert.pem -e BATON_PRIVATE_KEY=/out/sf-key.pem public.ecr.aws/conductorone/baton-successfactors:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-successfactors/cmd/baton-successfactors@main

baton-successfactors --public-key ./sf-cert.pem --private-key ./sf-key.pem

baton resources
```

# Data Model

`baton-successfactors` will pull down information about the following resources:
- Users
- Groups

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-successfactors` Command Line Usage

```
baton-successfactors

Usage:
  baton-successfactors [flags]
  baton-successfactors [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  config             Get the connector config schema
  health-check       Check the health of a running connector
  help               Help about any command

Flags:
      --auth-method string                               ($BATON_AUTH_METHOD)
      --cid string                                       required: Client ID ($BATON_CID)
      --client-id string                                 The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string                             The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --company-id string                                required: Company ID ($BATON_COMPANY_ID)
      --external-resource-c1z string                     The path to the c1z file to sync external baton resources with ($BATON_EXTERNAL_RESOURCE_C1Z)
      --external-resource-entitlement-id-filter string   The entitlement that external users, groups must have access to sync external baton resources ($BATON_EXTERNAL_RESOURCE_ENTITLEMENT_ID_FILTER)
      --external-resource-traits strings                 Resource type traits (e.g. "user", "group", "app") to sync and match from the external resource c1z. When unset the matcher falls back to user and group; passing this flag replaces the full set rather than adding to it. ($BATON_EXTERNAL_RESOURCE_TRAITS)
  -f, --file string                                      The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
      --health-check                                     Enable the HTTP health check endpoint ($BATON_HEALTH_CHECK)
      --health-check-port int                            Port for the HTTP health check endpoint ($BATON_HEALTH_CHECK_PORT) (default 8081)
  -h, --help                                             help for baton-successfactors
      --http-timeout-seconds int                         HTTP client timeout in seconds (max 1800) ($BATON_HTTP_TIMEOUT_SECONDS) (default 300)
      --instance-url string                              required: Your Success Factors domain, ex: https://successfactorsserver.com ($BATON_INSTANCE_URL)
      --issuer-url string                                required: Your SAML Issuer domain, ex: https://exampleissuer.com ($BATON_ISSUER_URL)
      --keep-previous-sync-c1z                           Keep the previously synced c1z on disk to enable ETag replay across service-mode syncs (requires a connector that supports ETag replay; costs one c1z of local disk) ($BATON_KEEP_PREVIOUS_SYNC_C1Z)
      --log-format string                                The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string                                 The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --log-level-debug-expires-at string                The timestamp indicating when debug-level logging should expire ($BATON_LOG_LEVEL_DEBUG_EXPIRES_AT)
      --log-path strings                                 The file path to write logs to ($BATON_LOG_PATH)
      --otel-collector-endpoint string                   The endpoint of the OpenTelemetry collector to send observability data to (used for both tracing and logging if specific endpoints are not provided) ($BATON_OTEL_COLLECTOR_ENDPOINT)
      --parallel-sync                                    Deprecated: use --workers instead. ($BATON_PARALLEL_SYNC)
      --private-key string                               required: Private key (PEM file) ($BATON_PRIVATE_KEY)
  -p, --provisioning                                     This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --public-key string                                required: Public key certificate (PEM file) ($BATON_PUBLIC_KEY)
      --saml-api-key string                              required: SAML API Key ($BATON_SAML_API_KEY)
      --skip-entitlements-and-grants                     This must be set to skip syncing of entitlements and grants ($BATON_SKIP_ENTITLEMENTS_AND_GRANTS)
      --skip-full-sync                                   This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --storage-engine string                            The storage engine to use when opening the sync c1z file. Only pebble is supported; defaults to pebble when unset. ($BATON_STORAGE_ENGINE)
      --subject-name-id string                           required: Subject Name ID ($BATON_SUBJECT_NAME_ID)
      --sync-resource-types strings                      The resource type IDs to sync ($BATON_SYNC_RESOURCE_TYPES)
      --sync-resources strings                           The resource IDs to sync ($BATON_SYNC_RESOURCES)
      --task-concurrency int                             The number of Baton tasks to run concurrently in service mode. Tasks may include sync, grant, revoke, and more. Minimum value is 1, maximum value is 100. ($BATON_TASK_CONCURRENCY) (default 3)
      --ticketing                                        This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                                          version for baton-successfactors
      --workers int                                      The number of sync workers to use. -1 for auto-detect, 0 for sequential, >0 for parallel ($BATON_WORKERS)

Use "baton-successfactors [command] --help" for more information about a command.
```
