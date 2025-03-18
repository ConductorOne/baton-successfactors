t#!/bin/bash

set -exo pipefail

 # CI test for use with CI PingFederate environment
if [ -z "$BATON_SUCCESSFACTORS" ]; then
  echo "BATON_SUCCESSFACTORS not set. using baton-successfactors"
  BATON_SUCCESSFACTORS=baton-successfactors
fi
if [ -z "$BATON" ]; then
  echo "BATON not set. using baton"
  BATON=baton
fi

# Error on unbound variables now that we've set BATON & BATON_SUCCESSFACTORS
set -u

# Sync
$BATON_SUCCESSFACTORS

# Check resources are synced
$BATON resources --output-format=json | jq ".resources[] | select(.resource.displayName == \"$BATON_PRINCIPAL\") | .resource"