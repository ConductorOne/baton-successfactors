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

	// The key fields come in pairs: PEM contents (a string) or a PEM file. Set exactly one of each
	// pair. The string fields keep the names and the meaning of earlier releases. The pair
	// constraints are in FieldRelationships, so no field is WithRequired.
	PubKeyField = field.StringField(
		"public-key",
		field.WithDisplayName("Certificate (PEM contents)"),
		field.WithDescription("Certificate as PEM text. Upload the PEM file in Certificate (PEM file) instead. "+
			"Use this field only if you cannot upload a file. Set one of the two fields."),
		field.WithPlaceholder("Your SuccessFactors certificate"),
		field.WithIsSecret(true),
	)

	PubKeyFileField = field.FileUploadField(
		"public-key-file",
		[]string{".pem", ".crt"},
		field.WithDisplayName("Certificate (PEM file)"),
		field.WithDescription("Upload the certificate PEM file (recommended). "+
			"Set this field or Certificate (PEM contents), not both."),
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

	PrivKeyField = field.StringField(
		"private-key",
		field.WithDisplayName("Private key (PEM contents)"),
		field.WithDescription("Private key as PEM text. Upload the PEM file in Private key (PEM file) instead. "+
			"Use this field only if you cannot upload a file. Set one of the two fields."),
		field.WithPlaceholder("Your SuccessFactors private key"),
		field.WithIsSecret(true),
	)

	PrivKeyFileField = field.FileUploadField(
		"private-key-file",
		[]string{".pem", ".key"},
		field.WithDisplayName("Private key (PEM file)"),
		field.WithDescription("Upload the private key PEM file (recommended). "+
			"Set this field or Private key (PEM contents), not both."),
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
