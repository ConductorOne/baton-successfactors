================
File: cmd/baton-successfactors/main.go
================
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/conductorone/baton-successfactors/pkg/connector"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-successfactors",
		getConnector,
		field.Configuration{
			Fields:           ConfigurationFields,
			FieldRelationships: FieldRelationships,
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	if err := ValidateConfig(ctx, v); err != nil {
		return nil, err
	}

	baseURL := v.GetString(BaseURLField.GetName())
	clientID := v.GetString(ClientIDField.GetName())
	companyID := v.GetString(CompanyIDField.GetName())
	certPrivateKey := v.GetString(CertPrivateKeyField.GetName())
	samlAssertionURL := v.GetString(SamlAssertionURLField.GetName())
	idpIssuerURL := v.GetString(IdpIssuerURLField.GetName())
	idpSubjectNameID := v.GetString(IdpSubjectNameIDField.GetName())

	sfClient, err := client.New(ctx, baseURL, clientID, companyID, certPrivateKey, samlAssertionURL, idpIssuerURL, idpSubjectNameID)
	if err != nil {
		l.Error("error creating client", zap.Error(err))
		return nil, err
	}

	cb, err := connector.New(ctx, sfClient)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	c, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return c, nil
}