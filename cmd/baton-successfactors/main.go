package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	cfg "github.com/conductorone/baton-successfactors/pkg/config"
	"github.com/conductorone/baton-successfactors/pkg/connector"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-successfactors",
		getConnector,
		cfg.Config,
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

func getConnector(ctx context.Context, config *cfg.Successfactors) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)
	if err := field.Validate(cfg.Config, config); err != nil {
		return nil, err
	}

	companyID := config.CompanyId
	if companyID == "" {
		return nil, fmt.Errorf("company-id is required")
	}

	clientID := config.Cid
	if clientID == "" {
		return nil, fmt.Errorf("cid is required")
	}

	certificate := config.PublicKey
	if certificate == "" {
		return nil, fmt.Errorf("public-key is required")
	}

	privateKey := config.PrivateKey
	if privateKey == "" {
		return nil, fmt.Errorf("private-key is required")
	}

	instanceURL := config.InstanceUrl
	if instanceURL == "" {
		return nil, fmt.Errorf("instance-url is required")
	}

	issuerURL := config.IssuerUrl
	if issuerURL == "" {
		return nil, fmt.Errorf("issuer-url is required")
	}

	username := config.SubjectNameId
	if username == "" {
		return nil, fmt.Errorf("subject-name-id is required")
	}

	samlAPIKey := config.SamlApiKey
	if samlAPIKey == "" {
		return nil, fmt.Errorf("saml-api-key is required")
	}

	cb, err := connector.New(
		ctx,
		companyID,
		clientID,
		certificate,
		privateKey,
		instanceURL,
		issuerURL,
		username,
		samlAPIKey,
	)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	conn, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return conn, nil
}
