package connector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/conductorone/baton-successfactors/internal/mocksf"
	cfg "github.com/conductorone/baton-successfactors/pkg/config"
)

func TestPemValue(t *testing.T) {
	assert.Equal(t, "text", pemValue("text", nil))
	assert.Equal(t, "file", pemValue("", []byte("file")))
	assert.Equal(t, "file", pemValue("text", []byte("file")), "the file wins when both are set")
	assert.Equal(t, "", pemValue("", nil))
	assert.Equal(t, "text", pemValue("text", []byte{}), "an empty file counts as not set")
}

func TestNewSelectsKeys(t *testing.T) {
	certPEM, keyPEM, err := mocksf.GenerateKeyPair(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	require.NoError(t, err)
	base := func() *cfg.Successfactors {
		return &cfg.Successfactors{
			CompanyId: "co", Cid: "cid", SamlApiKey: "key", InstanceUrl: "https://example.invalid",
			IssuerUrl: "https://issuer.example", SubjectNameId: "100001",
		}
	}
	ctx := context.Background()

	t.Run("text keys", func(t *testing.T) {
		c := base()
		c.PublicKey, c.PrivateKey = string(certPEM), string(keyPEM)
		_, _, err := New(ctx, c, nil)
		require.NoError(t, err)
	})
	t.Run("file keys", func(t *testing.T) {
		c := base()
		c.PublicKeyFile, c.PrivateKeyFile = certPEM, keyPEM
		_, _, err := New(ctx, c, nil)
		require.NoError(t, err)
	})
	t.Run("mixed", func(t *testing.T) {
		c := base()
		c.PublicKey, c.PrivateKeyFile = string(certPEM), keyPEM
		_, _, err := New(ctx, c, nil)
		require.NoError(t, err)
	})
	t.Run("missing certificate", func(t *testing.T) {
		c := base()
		c.PrivateKey = string(keyPEM)
		_, _, err := New(ctx, c, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "certificate is missing")
	})
	t.Run("missing private key", func(t *testing.T) {
		c := base()
		c.PublicKey = string(certPEM)
		_, _, err := New(ctx, c, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "private key is missing")
	})
	t.Run("invalid key", func(t *testing.T) {
		c := base()
		c.PublicKey, c.PrivateKey = "not a cert", "not a key"
		_, _, err := New(ctx, c, nil)
		require.Error(t, err)
	})
}
