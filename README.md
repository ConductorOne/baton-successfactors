![Baton Logo](./baton-logo.png)

# `baton-successfactors` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-successfactors.svg)](https://pkg.go.dev/github.com/conductorone/baton-successfactors) ![main ci](https://github.com/conductorone/baton-successfactors/actions/workflows/main.yaml/badge.svg)

`baton-successfactors` is a connector for SAP SuccessFactors built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It communicates with the OData V2 protocol to sync data about users.

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.
# Credentials
Please reference the [SAP SuccessFactors API Reference Guide- OData V2](https://help.sap.com/doc/a7c08a422cc14e1eaaffee83610a981d/2411/en-US/SF_HCM_OData_API_DEV.pdf) to configure the API credentials on SuccessFactors. To access SuccessFactor, you must provide the following:
1. SAML API key from when you register the connector as an OAuth application in SuccessFactors.
2. Issuer URL from when you register the connector as an OAuth application in SuccessFactors.
3. Client ID of the application from when you register the connector as an OAuth application in SuccessFactors.
4. X.509 Certificate for signing the SAML assertion from when you register the connector as an OAuth application in SuccessFactors
5. User ID of the admin user that will be added to the SAML subject. Check out [Person/User IDs Used Within Employee Central - Employee Profile](https://userapps.support.sap.com/sap/support/knowledge/en/2493579) to learn more.
6. Company ID for your company in SuccessFactors tenant. Follow [How to find the SAP SuccessFacotrs CompanyID](https://userapps.support.sap.com/sap/support/knowledge/en/2655655).
7. Instance URL for SuccessFactor API calls, you can find your API server following SAP documentation here [List of SAP SuccessFactors API server](https://help.sap.com/docs/SAP_SUCCESSFACTORS_PLATFORM/d599f15995d348a1b45ba5603e2aba9b/af2b8d5437494b12be88fe374eba75b6.html)

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-successfactors
baton-successfactors
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/conductorone/baton-successfactors:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-successfactors/cmd/baton-successfactors@main

baton-successfactors

baton resources
```

# Data Model

`baton-successfactors` will pull down information about the following resources:
- Users

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
  help               Help about any command

Flags:
      --cid string               required: Client ID ($BATON_CID)
      --client-id string         The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --company-id string        required: Company ID ($BATON_COMPANY_ID)
      --private-key string       required: Private Key ($BATON_PRIVATE_KEY)
      --public-key string        required: Public Key ($BATON_PUBLIC_KEY)
      --instance-url string      required: Your Success Factors domain, ex: https://successfactorsserver.com ($BATON_INSTANCE_URL)
      --issuer-url string        required: Your SAML Issuer domain, ex: https://exampleissuer.com ($BATON_ISSUER_URL)
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-successfactors
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-successfactors

Use "baton-successfactors [command] --help" for more information about a command.
```
