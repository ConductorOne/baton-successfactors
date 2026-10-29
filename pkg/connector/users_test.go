package connector

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/conductorone/baton-successfactors/pkg/connector/client"
)

func testKeyPair(t *testing.T) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return string(certPEM), string(keyPEM)
}

func TestUserBuilderListPagination(t *testing.T) {
	var mu sync.Mutex
	var empJobURIs []string

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth/token":
			_, _ = w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`))
		case "/odata/v2/EmpJob":
			mu.Lock()
			empJobURIs = append(empJobURIs, r.URL.RequestURI())
			mu.Unlock()
			body := client.SuccessFactorsUserIdList{Ds: client.D{
				Results: []client.Results{{UserId: "u1", UserNav: client.UserNav{Username: "user1", Email: "user1@example.com"}}},
				Next:    server.URL + "/odata/v2/EmpJob?$skiptoken=next-page",
			}}
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"d": body.Ds}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	cert, key := testKeyPair(t)
	c, err := client.New(ctx, server.URL, "company", "cid", cert, key, "https://issuer.example", "subject", "api-key")
	require.NoError(t, err)

	builder := newUserBuilder(c)

	// First page: empty token, the connector builds the initial query.
	resources, results, err := builder.List(ctx, nil, resource.SyncOpAttrs{})
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, "user1", resources[0].GetDisplayName())
	assert.Equal(t, server.URL+"/odata/v2/EmpJob?$skiptoken=next-page", results.NextPageToken)

	// Next page: the token from the previous page reaches the client as the request URL.
	_, _, err = builder.List(ctx, nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: results.NextPageToken}})
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, empJobURIs, 2)
	assert.Contains(t, empJobURIs[0], "%24expand=")
	assert.Equal(t, "/odata/v2/EmpJob?$skiptoken=next-page", empJobURIs[1])
}
