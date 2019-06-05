#!/bin/bash
./tools/bootstrap
mkdir -p $GOPATH/src/github.com/terraform-providers/terraform-provider-google
mkdir -p $GOPATH/src/github.com/terraform-providers/terraform-provider-google-beta
bundle exec compiler -a -v "ga" -e terraform -o "$GOPATH/src/github.com/terraform-providers/terraform-provider-google"
bundle exec compiler -a -v "beta" -e terraform -o "$GOPATH/src/github.com/terraform-providers/terraform-provider-google-beta"
