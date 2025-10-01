package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	CompIdField = field.StringField(
		"company-id",
		field.WithDisplayName("Company ID"),
		field.WithDescription("Company ID"),
		field.WithRequired(true),
	)

	ClientIdField = field.StringField(
		"cid",
		field.WithDisplayName("Client ID"),
		field.WithDescription("Client ID"),
		field.WithRequired(true),
	)

	PubKeyField = field.StringField(
		"public-key",
		field.WithDisplayName("Certificate (PEM)"),
		field.WithDescription("Public Key"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	SAMLAPIKeyField = field.StringField(
		"saml-api-key",
		field.WithDisplayName("SAML API key"),
		field.WithDescription("SAML API Key"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	PrivKeyField = field.StringField(
		"private-key",
		field.WithDisplayName("Private Key (PEM)"),
		field.WithDescription("Private Key"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	InstanceUrlField = field.StringField(
		"instance-url",
		field.WithDisplayName("Instance URL"),
		field.WithDescription("Your Success Factors domain, ex: https://successfactorsserver.com"),
		field.WithRequired(true),
	)

	IssuerUrlField = field.StringField(
		"issuer-url",
		field.WithDisplayName("Issuer URL"),
		field.WithDescription("Your SAML Issuer domain, ex: https://exampleissuer.com"),
		field.WithRequired(true),
	)

	SubjectNameIdField = field.StringField(
		"subject-name-id",
		field.WithDisplayName("Subject name ID"),
		field.WithDescription("Subject Name ID"),
		field.WithRequired(true),
	)

	ConfigurationFields = []field.SchemaField{
		CompIdField,
		ClientIdField,
		SAMLAPIKeyField,
		PubKeyField,
		PrivKeyField,
		InstanceUrlField,
		IssuerUrlField,
		SubjectNameIdField,
	}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

//go:generate go run -tags=generate ./gen
var Config = field.NewConfiguration(
	ConfigurationFields,
	field.WithConstraints(FieldRelationships...),
	field.WithConnectorDisplayName("SuccessFactors"),
	field.WithHelpUrl("/docs/baton/successfactors"),
	field.WithIconUrl("/static/app-icons/successfactors.svg"),
)
