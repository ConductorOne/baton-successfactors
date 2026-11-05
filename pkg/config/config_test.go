package config

import (
	"testing"

	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/stretchr/testify/assert"
)

// base returns a config with all fields set except the four key fields.
func base() *Successfactors {
	return &Successfactors{
		CompanyId:     "company_id",
		Cid:           "c_id",
		SamlApiKey:    "abcdef12345",
		InstanceUrl:   "https://test",
		IssuerUrl:     "https://test",
		SubjectNameId: "user",
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *Successfactors)
		wantErr bool
	}{
		{"text keys only", func(c *Successfactors) { c.PublicKey, c.PrivateKey = "qwertyuiop", "asdfghjkl" }, false},
		{"file keys only", func(c *Successfactors) { c.PublicKeyFile, c.PrivateKeyFile = []byte("qwertyuiop"), []byte("asdfghjkl") }, false},
		{"mixed: text certificate, file key", func(c *Successfactors) { c.PublicKey, c.PrivateKeyFile = "qwertyuiop", []byte("asdfghjkl") }, false},
		{"both certificate fields set", func(c *Successfactors) {
			c.PublicKey, c.PublicKeyFile, c.PrivateKey = "qwertyuiop", []byte("qwertyuiop"), "asdfghjkl"
		}, true},
		{"both private key fields set", func(c *Successfactors) {
			c.PublicKey, c.PrivateKey, c.PrivateKeyFile = "qwertyuiop", "asdfghjkl", []byte("asdfghjkl")
		}, true},
		{"no certificate", func(c *Successfactors) { c.PrivateKey = "asdfghjkl" }, true},
		{"no private key", func(c *Successfactors) { c.PublicKey = "qwertyuiop" }, true},
		{"neither key set", func(c *Successfactors) {}, true},
		{"missing required fields", func(c *Successfactors) {
			*c = Successfactors{Cid: "c_id", SamlApiKey: "abcdef12345", PublicKey: "qwertyuiop", PrivateKey: "asdfghjkl"}
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := base()
			tt.mutate(c)
			err := field.Validate(Config, c)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
