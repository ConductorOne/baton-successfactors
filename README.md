![Baton Logo](./baton-logo.png)

# `baton-successfactors` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-successfactors.svg)](https://pkg.go.dev/github.com/conductorone/baton-successfactors) ![main ci](https://github.com/conductorone/baton-successfactors/actions/workflows/main.yaml/badge.svg)

`baton-successfactors` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

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
      --client-secret string     The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --company-id string        required: Company ID ($BATON_COMPANY_ID)
  -f, --file string              The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                     help for baton-successfactors
      --instance-url string      required: Your Success Factors domain, ex: https://successfactorsserver.com ($BATON_INSTANCE_URL)
      --issuer-url string        required: Your SAML Issuer domain, ex: https://exampleissuer.com ($BATON_ISSUER_URL)
      --log-format string        The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string         The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --private-key string       required: Private Key ($BATON_PRIVATE_KEY)
  -p, --provisioning             This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --public-key string        required: Public Key ($BATON_PUBLIC_KEY)
      --saml-api-key string      required: SAML API Key ($BATON_SAML_API_KEY)
      --skip-full-sync           This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --subject-name-id string   required: Subject Name ID ($BATON_SUBJECT_NAME_ID)
      --ticketing                This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                  version for baton-successfactors

Use "baton-successfactors [command] --help" for more information about a command.
```
