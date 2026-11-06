package connector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
func newMockClient(t *testing.T, users []json.RawMessage, mutate ...func(*mocksf.Config)) (*client.SuccessFactorsClient, *httptest.Server, func() []string) {
	t.Helper()
	return newInterceptedMockClient(t, users, nil, mutate...)
}

// newInterceptedMockClient is newMockClient with a hook that can answer a request before the mock does.
// The hook returns true when it has answered.
func newInterceptedMockClient(
	t *testing.T,
	users []json.RawMessage,
	intercept func(http.ResponseWriter, *http.Request) bool,
	mutate ...func(*mocksf.Config),
) (*client.SuccessFactorsClient, *httptest.Server, func() []string) {
	t.Helper()
	certPEM, keyPEM, err := mocksf.GenerateKeyPair(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	require.NoError(t, err)
	cert, err := mocksf.ParseCertPEM(certPEM)
	require.NoError(t, err)

	cfg := mocksf.Config{
		Company: testCompany, ClientID: testClient, APIKey: testAPIKey, Issuer: testIssuer, Subject: testSubject,
		Cert: cert, Users: users, PageSize: 5, Log: &strings.Builder{},
	}
	for _, m := range mutate {
		m(&cfg)
	}
	h, err := mocksf.NewHandler(cfg)
	require.NoError(t, err)

	var mu sync.Mutex
	var empJobURIs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if intercept != nil && intercept(w, r) {
			return
		}
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

func failPage(page, httpStatus int) func(*mocksf.Config) {
	return func(c *mocksf.Config) { c.FailPage, c.FailStatus = page, httpStatus }
}

// A 404 from the user list must fail the sync with a code that the SDK neither
// skips (NotFound) nor retries without limit (Unavailable, DeadlineExceeded).
func requireFailedPrecondition(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	code := status.Code(err)
	assert.Equal(t, codes.FailedPrecondition, code)
	assert.NotContains(t, []codes.Code{codes.NotFound, codes.Unavailable, codes.DeadlineExceeded}, code)
	assert.Contains(t, err.Error(), "user list returned not found")
}

func TestUserBuilderListNotFoundOnLaterPageFails(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)
	c, _, _ := newMockClient(t, users, failPage(2, http.StatusNotFound))
	builder := newUserBuilder(c)

	_, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{})
	require.NoError(t, err)
	require.NotEmpty(t, results.NextPageToken)

	_, _, err = builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: results.NextPageToken}})
	requireFailedPrecondition(t, err)
	assert.Contains(t, err.Error(), "page token set: true")
	assert.NotContains(t, err.Error(), results.NextPageToken, "the page token is a URL and must not be printed")
}

func TestUserBuilderListNotFoundOnFirstPageFails(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)
	c, _, _ := newMockClient(t, users, failPage(1, http.StatusNotFound))

	_, _, err = newUserBuilder(c).List(context.Background(), nil, resource.SyncOpAttrs{})
	requireFailedPrecondition(t, err)
	assert.Contains(t, err.Error(), "page token set: false")
}

func TestUserBuilderListOtherErrorsUnchanged(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)
	c, _, _ := newMockClient(t, users, failPage(2, http.StatusBadRequest))
	builder := newUserBuilder(c)

	_, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{})
	require.NoError(t, err)
	_, _, err = builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: results.NextPageToken}})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, err.Error(), "failed to list users")
}

func TestRedactPageToken(t *testing.T) {
	skip := "eyJzdGFydFJvdyI6NSwiZW5kUm93IjoxMH0="
	token := "http://127.0.0.1:8089/odata/v2/EmpJob?" + url.Values{"$skiptoken": {skip}}.Encode()

	tests := []struct {
		name  string
		msg   string
		token string
		want  string
	}{
		{"empty token", "404 Not Found: " + token, "", "404 Not Found: " + token},
		{"raw token", "404: " + token + " end", token, "404: <page-token> end"},
		{"query-escaped token", "404: url=" + url.QueryEscape(token), token, "404: url=<page-token>"},
		{"path-escaped token", "404: " + url.PathEscape(token), token, "404: <page-token>"},
		{"skiptoken value alone", "bad skiptoken " + skip + " here", token, "bad skiptoken <redacted> here"},
		{"escaped skiptoken value alone", "bad " + url.QueryEscape(skip), token, "bad <redacted>"},
		{"no token in the message", "404 Not Found", token, "404 Not Found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, redactPageToken(tt.msg, tt.token))
		})
	}
}

func TestUserBuilderListNotFoundErrorHasNoPageToken(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)

	// The server answers a page request with a 404 that echoes the full request URL in its body.
	// Today the client error does not carry the response body, so this test also passes without
	// redactPageToken. It guards against the body or the URL reaching the error text later.
	var server *httptest.Server
	echo404 := func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != mocksf.EmpJobPath || r.URL.Query().Get("$skiptoken") == "" {
			return false
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"errorHttpCode": "404",
			"errorMessage":  "Page not found: " + server.URL + r.URL.RequestURI(),
		})
		return true
	}
	c, srv, _ := newInterceptedMockClient(t, users, echo404)
	server = srv
	builder := newUserBuilder(c)

	_, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{})
	require.NoError(t, err)
	token := results.NextPageToken
	require.NotEmpty(t, token)
	u, err := url.Parse(token)
	require.NoError(t, err)
	skip := u.Query().Get("$skiptoken")
	require.NotEmpty(t, skip)

	_, _, err = builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: token}})
	requireFailedPrecondition(t, err)
	assert.NotContains(t, err.Error(), token)
	assert.NotContains(t, err.Error(), skip)
	assert.NotContains(t, err.Error(), url.QueryEscape(skip))
	assert.Contains(t, err.Error(), "page token set: true")
}

func TestUserBuilderListShowsSuccessFactorsMessages(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)

	t.Run("wrong company shows the token error message", func(t *testing.T) {
		// The mock knows another company than the client sends.
		c, _, _ := newMockClient(t, users, func(cfg *mocksf.Config) { cfg.Company = "OTHERCO" })
		_, _, err := newUserBuilder(c).List(context.Background(), nil, resource.SyncOpAttrs{})
		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
		assert.Contains(t, err.Error(), "Company "+testCompany+" does not exist in this data center")
	})

	t.Run("400 on page 2 shows the OData message", func(t *testing.T) {
		c, _, _ := newMockClient(t, users, failPage(2, http.StatusBadRequest))
		builder := newUserBuilder(c)
		_, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{})
		require.NoError(t, err)
		_, _, err = builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: results.NextPageToken}})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "INJECTED_FAILURE: Synthetic: injected failure.")
	})

	t.Run("503 is still Unavailable", func(t *testing.T) {
		c, _, _ := newMockClient(t, users, failPage(2, http.StatusServiceUnavailable))
		builder := newUserBuilder(c)
		_, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{})
		require.NoError(t, err)
		_, _, err = builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: results.NextPageToken}})
		require.Error(t, err)
		assert.Equal(t, codes.Unavailable, status.Code(err), "the syncer retries Unavailable, as before")
	})

	t.Run("400 that echoes the page URL has no page token in the error", func(t *testing.T) {
		var server *httptest.Server
		echo400 := func(w http.ResponseWriter, r *http.Request) bool {
			if r.URL.Path != mocksf.EmpJobPath || r.URL.Query().Get("$skiptoken") == "" {
				return false
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{
				"code":    "BAD_REQUEST",
				"message": map[string]string{"lang": "en-US", "value": "Invalid page " + server.URL + r.URL.RequestURI()},
			}})
			return true
		}
		c, srv, _ := newInterceptedMockClient(t, users, echo400)
		server = srv
		builder := newUserBuilder(c)

		_, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{})
		require.NoError(t, err)
		token := results.NextPageToken
		u, err := url.Parse(token)
		require.NoError(t, err)
		skip := u.Query().Get("$skiptoken")
		require.NotEmpty(t, skip)

		_, _, err = builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: token}})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "Invalid page", "the SAP message is shown")
		assert.NotContains(t, err.Error(), token)
		assert.NotContains(t, err.Error(), skip)
		assert.NotContains(t, err.Error(), url.QueryEscape(skip))
	})
}

func TestUserBuilderListKeepsRateLimitDetails(t *testing.T) {
	users, err := mocksf.DefaultUsers()
	require.NoError(t, err)

	// The server answers a page request with a 429 and an OData body that echoes the request URL.
	var server *httptest.Server
	limited := func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != mocksf.EmpJobPath || r.URL.Query().Get("$skiptoken") == "" {
			return false
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Ratelimit-Limit", "100")
		w.Header().Set("X-Ratelimit-Remaining", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{
			"code":    "TOO_MANY_REQUESTS",
			"message": map[string]string{"lang": "en-US", "value": "Slow down " + server.URL + r.URL.RequestURI()},
		}})
		return true
	}
	c, srv, _ := newInterceptedMockClient(t, users, limited)
	server = srv
	builder := newUserBuilder(c)

	_, results, err := builder.List(context.Background(), nil, resource.SyncOpAttrs{})
	require.NoError(t, err)
	token := results.NextPageToken
	u, err := url.Parse(token)
	require.NoError(t, err)
	skip := u.Query().Get("$skiptoken")
	require.NotEmpty(t, skip)

	_, _, err = builder.List(context.Background(), nil, resource.SyncOpAttrs{PageToken: pagination.Token{Token: token}})
	require.Error(t, err)
	assert.Equal(t, codes.Unavailable, status.Code(err), "the syncer retries Unavailable")

	// The SDK retry reads the rate limit data from the status details.
	var found *v2.RateLimitDescription
	for _, d := range status.Convert(err).Details() {
		if rl, ok := d.(*v2.RateLimitDescription); ok {
			found = rl
		}
	}
	require.NotNil(t, found, "the rate limit details must survive the redaction")
	assert.Equal(t, v2.RateLimitDescription_STATUS_OVERLIMIT, found.GetStatus())

	assert.Contains(t, err.Error(), "Slow down", "the SAP message is shown")
	assert.NotContains(t, err.Error(), token)
	assert.NotContains(t, err.Error(), skip)
	assert.NotContains(t, err.Error(), url.QueryEscape(skip))
}

func TestListError(t *testing.T) {
	token := "http://127.0.0.1:8089/odata/v2/EmpJob?" + url.Values{"$skiptoken": {"eyJzdGFydFJvdyI6NX0="}}.Encode()

	rl, err := status.New(codes.Unavailable, "429 Too Many Requests").WithDetails(
		v2.RateLimitDescription_builder{Status: v2.RateLimitDescription_STATUS_OVERLIMIT}.Build())
	require.NoError(t, err)

	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{"deadline exceeded", fmt.Errorf("get %s: %w", token, context.DeadlineExceeded), codes.DeadlineExceeded},
		{"canceled", fmt.Errorf("get %s: %w", token, context.Canceled), codes.Canceled},
		{"plain error", errors.New("parse " + token + ": bad"), codes.Unknown},
		{"status error with details", fmt.Errorf("%s: %w", token, rl.Err()), codes.Unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listError(tt.err, token)
			require.Error(t, got)
			assert.Equal(t, tt.want, status.Code(got))
			assert.NotContains(t, got.Error(), token)
			assert.NotContains(t, got.Error(), "eyJzdGFydFJvdyI6NX0")
			assert.Contains(t, got.Error(), "baton-successfactors: failed to list users")
		})
	}

	t.Run("rate limit details are kept", func(t *testing.T) {
		got := listError(rl.Err(), token)
		var found bool
		for _, d := range status.Convert(got).Details() {
			_, ok := d.(*v2.RateLimitDescription)
			found = found || ok
		}
		assert.True(t, found)
	})
}
