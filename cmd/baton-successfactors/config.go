package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

var (
	BaseURLField = field.StringField(
		"base-url",
		field.WithDescription("The base URL for the SAP SuccessFactors OData API (e.g., https://api17.sapsf.com)."),
		field.WithRequired(true),
	)

	ClientIDField = field.StringField(
		"client-id",
		field.WithDescription("The API key obtained after registering the OAuth 2.0 client application."),
		field.WithRequired(true),
	)

	CompanyIDField = field.StringField(
		"company-id",
		field.WithDescription("Your SAP SuccessFactors company ID."),
		field.WithRequired(true),
	)

	CertPrivateKeyField = field.StringField(
		"cert-private-key",
		field.WithDescription("The X.509 certificate private key used to sign the SAML assertion."),
		field.WithRequired(true),
	)

	SamlAssertionURLField = field.StringField(
		"saml-assertion-url",
		field.WithDescription("The url to use when requesting the oAuth token"),
		field.WithDefaultValue(""),
	)

	IdpIssuerURLField = field.StringField(
		"idp-issuer-url",
		field.WithDescription("The Issuer URL value for the IdP used to create the SAML assertion."),
	)

	IdpSubjectNameIDField = field.StringField(
		"idp-subject-nameid",
		field.WithDescription("The Subject Name ID value for the user used to create the SAML assertion."),
	)

	ConfigurationFields = []field.SchemaField{
		BaseURLField,
		ClientIDField,
		CompanyIDField,
		CertPrivateKeyField,
		SamlAssertionURLField,
		IdpIssuerURLField,
		IdpSubjectNameIDField,
	}

	FieldRelationships = []field.SchemaFieldRelationship{}
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(ctx context.Context, v *viper.Viper) error {
	if v.GetString(BaseURLField.GetName()) == "" {
		return fmt.Errorf("base URL is required")
	}

	if v.GetString(ClientIDField.GetName()) == "" {
		return fmt.Errorf("client ID is required")
	}

	if v.GetString(CompanyIDField.GetName()) == "" {
		return fmt.Errorf("company ID is required")
	}

	if v.GetString(CertPrivateKeyField.GetName()) == "" {
		return fmt.Errorf("certificate private key is required")
	}

	return nil
}