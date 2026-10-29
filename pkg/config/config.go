package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	CompIdField = field.StringField(
		"company-id",
		field.WithDisplayName("Company ID"),
		field.WithDescription("Company ID"),
		field.WithPlaceholder("Your SuccessFactors Company ID"),
		field.WithRequired(true),
	)

	ClientIdField = field.StringField(
		"cid",
		field.WithDisplayName("Client ID"),
		field.WithDescription("Client ID"),
		field.WithPlaceholder("Your SuccessFactors Client ID"),
		field.WithRequired(true),
	)

	PubKeyField = field.FileUploadField(
		"public-key",
		[]string{".pem", ".crt"},
		field.WithDisplayName("Certificate (PEM)"),
		field.WithDescription("Public key certificate (PEM file)"),
		field.WithPlaceholder("Your SuccessFactors certificate"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	SAMLAPIKeyField = field.StringField(
		"saml-api-key",
		field.WithDisplayName("SAML API key"),
		field.WithDescription("SAML API Key"),
		field.WithPlaceholder("Your SuccessFactors SAML API key"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	PrivKeyField = field.FileUploadField(
		"private-key",
		[]string{".pem", ".key"},
		field.WithDisplayName("Private Key (PEM)"),
		field.WithDescription("Private key (PEM file)"),
		field.WithPlaceholder("Your SuccessFactors private key"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	InstanceUrlField = field.StringField(
		"instance-url",
		field.WithDisplayName("Instance URL"),
		field.WithDescription("Your Success Factors domain, ex: https://successfactorsserver.com"),
		field.WithPlaceholder("Your SuccessFactors instance URL"),
		field.WithRequired(true),
	)

	IssuerUrlField = field.StringField(
		"issuer-url",
		field.WithDisplayName("Issuer URL"),
		field.WithDescription("Your SAML Issuer domain, ex: https://exampleissuer.com"),
		field.WithPlaceholder("Your SuccessFactors issuer URL"),
		field.WithRequired(true),
	)

	SubjectNameIdField = field.StringField(
		"subject-name-id",
		field.WithDisplayName("Subject name ID"),
		field.WithDescription("Subject Name ID"),
		field.WithPlaceholder("Your SuccessFactors subject name ID"),
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
