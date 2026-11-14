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

	// The key fields come in pairs: PEM contents (a string) or a PEM file. Set exactly one of each
	// pair. The string fields keep the names and the meaning of earlier releases. They are CLI-only:
	// the C1 UI shows only the file fields. C1 maps saved public-key/private-key values to the file
	// fields (FieldNameMapping in C1). The pair constraints are in FieldRelationships, so no field is
	// WithRequired. baton-sdk leaves a constraint out of config_schema.json when it includes a
	// CLI-only field, but it still enforces the constraint at runtime (CLI and Lambda).
	PubKeyField = field.StringField(
		"public-key",
		field.WithDisplayName("Certificate (PEM contents)"),
		field.WithDescription("Certificate as PEM text (service mode and CLI only). "+
			"Use public-key-file to pass a PEM file path instead. Set one of the two."),
		field.WithIsSecret(true),
		field.WithExportTarget(field.ExportTargetCLIOnly),
	)

	PubKeyFileField = field.FileUploadField(
		"public-key-file",
		[]string{".pem", ".crt"},
		field.WithDisplayName("Certificate (PEM)"),
		field.WithDescription("Upload the certificate PEM file. In service mode or the CLI, pass the path to the file."),
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
		field.WithDisplayName("Private key (PEM contents)"),
		field.WithDescription("Private key as PEM text (service mode and CLI only). "+
			"Use private-key-file to pass a PEM file path instead. Set one of the two."),
		field.WithIsSecret(true),
		field.WithExportTarget(field.ExportTargetCLIOnly),
	)

	PrivKeyFileField = field.FileUploadField(
		"private-key-file",
		[]string{".pem", ".key"},
		field.WithDisplayName("Private key (PEM)"),
		field.WithDescription("Upload the private key PEM file. In service mode or the CLI, pass the path to the file."),
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
		PubKeyFileField,
		PrivKeyField,
		PrivKeyFileField,
		InstanceUrlField,
		IssuerUrlField,
		SubjectNameIdField,
	}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{
		field.FieldsMutuallyExclusive(PubKeyField, PubKeyFileField),
		field.FieldsAtLeastOneUsed(PubKeyField, PubKeyFileField),
		field.FieldsMutuallyExclusive(PrivKeyField, PrivKeyFileField),
		field.FieldsAtLeastOneUsed(PrivKeyField, PrivKeyFileField),
	}
)

//go:generate go run -tags=generate ./gen
var Config = field.NewConfiguration(
	ConfigurationFields,
	field.WithConstraints(FieldRelationships...),
	field.WithConnectorDisplayName("SuccessFactors"),
	field.WithHelpUrl("/docs/baton/successfactors"),
	field.WithIconUrl("/static/app-icons/successfactors.svg"),
)
