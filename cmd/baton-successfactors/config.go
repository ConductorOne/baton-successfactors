package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

var (
	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	CompIdField = field.StringField(
		"company-id",
		field.WithDescription("Company ID"),
		field.WithRequired(true),
	)
	ClientIdField = field.StringField(
		"cid",
		field.WithDescription("Client ID"),
		field.WithRequired(true),
	)
	PubKeyField = field.StringField(
		"public-key",
		field.WithDescription("Public Key"),
		field.WithRequired(true),
	)
	SAMLAPIKeyField = field.StringField(
		"saml-api-key",
		field.WithDescription("SAML API Key"),
		field.WithRequired(true),
	)
	PrivKeyField = field.StringField(
		"private-key",
		field.WithDescription("Private Key"),
		field.WithRequired(true),
	)
	InstanceUrlField = field.StringField(
		"instance-url",
		field.WithDescription("Your Success Factors domain, ex: https://successfactorsserver.com"),
		field.WithRequired(true),
	)
	IssuerUrlField = field.StringField(
		"issuer-url",
		field.WithDescription("Your SAML Issuer domain, ex: https://exampleissuer.com"),
		field.WithRequired(true),
	)
	SubjectNameIdField = field.StringField(
		"subject-name-id",
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

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(v *viper.Viper) error {
	return nil
}
