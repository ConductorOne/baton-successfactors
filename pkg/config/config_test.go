package config

import (
	"testing"

	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Successfactors
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Successfactors{
				CompanyId:     "company_id",
				Cid:           "c_id",
				SamlApiKey:    "abcdef12345",
				PublicKey:     "qwertyuiop",
				PrivateKey:    "asdfghjkl",
				InstanceUrl:   "https://test",
				IssuerUrl:     "https://test",
				SubjectNameId: "user",
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing required fields",
			config: &Successfactors{
				Cid:        "c_id",
				SamlApiKey: "abcdef12345",
				PublicKey:  "qwertyuiop",
				PrivateKey: "asdfghjkl",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := field.Validate(Config, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
