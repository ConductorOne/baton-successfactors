package connector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/conductorone/baton-successfactors/internal/mocksf"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
)

const (
	testCompany = "company"
	testClient  = "cid"
	testIssuer  = "https://issuer.example"
	testSubject = "subject"
	testAPIKey  = "api-key"
)

// newMockClient starts the mock SuccessFactors server and returns a client for it.
// It returns the request URIs that the mock gets on /odata/v2/EmpJob.
func newMockClient(t *testing.T, users []json.RawMessage) (*client.SuccessFactorsClient, *httptest.Server, func() []string) {
	t.Helper()
	certPEM, keyPEM, err := mocksf.GenerateKeyPair(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	require.NoError(t, err)
	cert, err := mocksf.ParseCertPEM(certPEM)
	require.NoError(t, err)

	h, err := mocksf.NewHandler(mocksf.Config{
		Company: testCompany, ClientID: testClient, APIKey: testAPIKey, Issuer: testIssuer, Subject: testSubject,
		Cert: cert, Users: users, PageSize: 5, Log: &strings.Builder{},
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var empJobURIs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mocksf.EmpJobPath {
			mu.Lock()
			empJobURIs = append(empJobURIs, r.URL.RequestURI())
			mu.Unlock()
		}
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)

	c, err := client.New(context.Background(), server.URL, testCompany, testClient, string(certPEM), string(keyPEM), testIssuer, testSubject, testAPIKey)
	require.NoError(t, err)
	return c, server, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), empJobURIs...)
	}
}

type expectedUsers struct {
	Count int `json:"count"`
	Users map[string]struct {
		Status  string            `json:"status"`
		Profile map[string]string `json:"profile"`
	} `json:"users"`
}

// listAll follows the page tokens like the syncer does.
func listAll(t *testing.T, builder *userBuilder) ([]*v2.Resource, []int) {
	t.Helper()
	var all []*v2.Resource
	var sizes []int
	token := ""
	for i := 0; i < 20; i++ {
		resources, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: token}})
		require.NoError(t, err)
		all = append(all, resources...)
		sizes = append(sizes, len(resources))
		token = results.NextPageToken
		if token == "" {
			return all, sizes
		}
	}
	t.Fatal("the page tokens did not end")
	return nil, nil
}

func TestUserBuilderListPagination(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)
	c, server, requests := newMockClient(t, users)
	builder := newUserBuilder(c)

	// First page: empty token, the connector builds the initial query.
	resources, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{})
	require.NoError(t, err)
	require.Len(t, resources, 5)
	assert.True(t, strings.HasPrefix(results.NextPageToken, server.URL+mocksf.EmpJobPath+"?%24skiptoken="), results.NextPageToken)

	// Next page: the token from the previous page reaches the client as the request URL.
	_, _, err = builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: results.NextPageToken}})
	require.NoError(t, err)

	uris := requests()
	require.Len(t, uris, 2)
	assert.Contains(t, uris[0], "%24expand=")
	assert.Equal(t, strings.TrimPrefix(results.NextPageToken, server.URL), uris[1])
}

func TestUserBuilderListAllPages(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)
	c, _, _ := newMockClient(t, users)

	all, sizes := listAll(t, newUserBuilder(c))
	assert.Equal(t, []int{5, 5, 2}, sizes, "the last page has no next token")
	assert.Len(t, all, len(users))
}

func TestUserBuilderMapsFixtureUsers(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)
	c, _, _ := newMockClient(t, users)
	all, _ := listAll(t, newUserBuilder(c))

	var exp expectedUsers
	require.NoError(t, json.Unmarshal(mocksf.ExpectedJSON, &exp))
	require.Len(t, all, exp.Count)

	wantStatus := map[string]v2.Status_ResourceStatus{"enabled": v2.Status_RESOURCE_STATUS_ENABLED, "disabled": v2.Status_RESOURCE_STATUS_DISABLED}
	seen := map[string]bool{}
	for _, r := range all {
		name := r.GetDisplayName()
		seen[name] = true
		want, ok := exp.Users[name]
		require.True(t, ok, "unexpected user %q", name)
		assert.Equal(t, wantStatus[want.Status], r.GetStatus().GetStatus(), "status of %s", name)
		profile := r.GetProfile().AsMap()
		for k, v := range want.Profile {
			assert.Equal(t, v, profile[k], "profile %s of %s", k, name)
		}
	}
	assert.Len(t, seen, exp.Count)
}

func TestUserBuilderBadDateFails(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)
	users, err = mocksf.ApplyScenario(users, mocksf.ScenarioBadDate)
	require.NoError(t, err)
	c, _, _ := newMockClient(t, users)
	builder := newUserBuilder(c)

	var lastErr error
	token := ""
	for i := 0; i < 10 && lastErr == nil; i++ {
		_, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: token}})
		lastErr = err
		if err == nil {
			if results.NextPageToken == "" {
				break
			}
			token = results.NextPageToken
		}
	}
	require.Error(t, lastErr)
	assert.Contains(t, lastErr.Error(), "failed to build user resource")
}
