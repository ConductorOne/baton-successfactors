package mocksf

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/conductorone/baton-successfactors/pkg/connector/client"
)

type env struct {
	srv     *httptest.Server
	cfg     Config
	certPEM string
	keyPEM  string
	log     *bytes.Buffer
}

func newEnv(t *testing.T, notBefore, notAfter time.Time, mutate func(*Config)) *env {
	t.Helper()
	certPEM, keyPEM, err := GenerateKeyPair(notBefore, notAfter)
	require.NoError(t, err)
	cert, err := ParseCertPEM(certPEM)
	require.NoError(t, err)
	users, err := DefaultUsers()
	require.NoError(t, err)
	e := &env{log: &bytes.Buffer{}, certPEM: string(certPEM), keyPEM: string(keyPEM)}
	e.cfg = Config{
		Company: "co", ClientID: "cid", APIKey: "key", Issuer: "https://issuer.example", Subject: "100001",
		Cert: cert, Users: users, Log: e.log,
	}
	if mutate != nil {
		mutate(&e.cfg)
	}
	h, err := NewHandler(e.cfg)
	require.NoError(t, err)
	e.srv = httptest.NewServer(h)
	t.Cleanup(e.srv.Close)
	return e
}

func validEnv(t *testing.T, mutate func(*Config)) *env {
	return newEnv(t, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), mutate)
}

// client builds the real connector client against the mock.
func (e *env) client(t *testing.T, company, subject string) *client.SuccessFactorsClient {
	t.Helper()
	c, err := client.New(context.Background(), e.srv.URL, company, "cid", e.certPEM, e.keyPEM, "https://issuer.example", subject, "key")
	require.NoError(t, err)
	return c
}

func post(t *testing.T, target string, form url.Values) (int, string) {
	t.Helper()
	resp, err := http.Post(target, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func TestTokenURLParameterRejected(t *testing.T) {
	e := validEnv(t, nil)
	status, body := post(t, e.srv.URL+TokenPath+"?assertion=x", url.Values{})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.JSONEq(t, `{"errorHttpCode":"400","errorMessage":"SAML assertion is not allowed in URL parameter."}`, body)
}

func TestTokenWrongContentType(t *testing.T) {
	e := validEnv(t, nil)
	resp, err := http.Post(e.srv.URL+TokenPath, "application/json", strings.NewReader("{}"))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSyncPagesWithRealClient(t *testing.T) {
	e := validEnv(t, func(c *Config) { c.PageSize = 5 })
	c := e.client(t, "co", "100001")

	var sizes []int
	token := ""
	for i := 0; i < 10; i++ {
		users, next, err := c.GetUserData(context.Background(), token)
		require.NoError(t, err)
		sizes = append(sizes, len(users))
		if next == "" {
			break
		}
		assert.True(t, strings.HasPrefix(next, e.srv.URL+EmpJobPath+"?%24skiptoken="), next)
		token = next
	}
	assert.Equal(t, []int{5, 5, 2}, sizes)
}

func TestSkipTokenFormat(t *testing.T) {
	e := validEnv(t, nil)
	c := e.client(t, "co", "100001")
	_, next, err := c.GetUserData(context.Background(), "")
	require.NoError(t, err)
	u, err := url.Parse(next)
	require.NoError(t, err)
	raw, err := base64.StdEncoding.DecodeString(u.Query().Get("$skiptoken"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"startRow":5,"endRow":10}`, string(raw))
}

func TestWrongCompany(t *testing.T) {
	e := validEnv(t, nil)
	c := e.client(t, "other", "100001")
	_, _, err := c.GetUserData(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestCompanyBodyMatchesRealFormat(t *testing.T) {
	e := validEnv(t, nil)
	status, body := post(t, e.srv.URL+TokenPath, url.Values{"company_id": {"other"}})
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.JSONEq(t, `{"errorHttpCode":"401","errorMessage":"Company other does not exist in this data center. Please verify the company id and the endpoint URL you are using."}`, body)
}

func TestWrongSubjectIssuerKeyAndClient(t *testing.T) {
	e := validEnv(t, nil)
	_, _, err := e.client(t, "co", "wrong-subject").GetUserData(context.Background(), "")
	require.Error(t, err)

	// A different cert than the mock trusts gives a signature failure.
	other, otherKey, err := GenerateKeyPair(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	require.NoError(t, err)
	c, err := client.New(context.Background(), e.srv.URL, "co", "cid", string(other), string(otherKey), "https://issuer.example", "100001", "key")
	require.NoError(t, err)
	_, _, err = c.GetUserData(context.Background(), "")
	require.Error(t, err)

	c, err = client.New(context.Background(), e.srv.URL, "co", "wrong-client", e.certPEM, e.keyPEM, "https://issuer.example", "100001", "key")
	require.NoError(t, err)
	_, _, err = c.GetUserData(context.Background(), "")
	require.Error(t, err)
}

func TestExpiredAssertionRejected(t *testing.T) {
	// The assertion window starts now. A mock clock far in the future is outside it.
	e := validEnv(t, func(c *Config) { c.Now = func() time.Time { return time.Now().Add(48 * time.Hour) } })
	_, _, err := e.client(t, "co", "100001").GetUserData(context.Background(), "")
	require.Error(t, err)
}

func TestExpiredCertToggle(t *testing.T) {
	past := time.Now().Add(-48 * time.Hour)
	accept := newEnv(t, past.Add(-time.Hour), past, nil)
	users, _, err := accept.client(t, "co", "100001").GetUserData(context.Background(), "")
	require.NoError(t, err)
	assert.NotEmpty(t, users)

	reject := newEnv(t, past.Add(-time.Hour), past, func(c *Config) { c.RejectExpiredCert = true })
	_, _, err = reject.client(t, "co", "100001").GetUserData(context.Background(), "")
	require.Error(t, err)
}

func TestEmpJobRequiresBearer(t *testing.T) {
	e := validEnv(t, nil)
	resp, err := http.Get(e.srv.URL + EmpJobPath)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	req, _ := http.NewRequest(http.MethodGet, e.srv.URL+EmpJobPath, nil)
	req.Header.Set("Authorization", "Bearer made-up")
	resp2, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode)
}

func TestFailPage(t *testing.T) {
	e := validEnv(t, func(c *Config) { c.FailPage = 2; c.FailStatus = 500 })
	c := e.client(t, "co", "100001")
	_, next, err := c.GetUserData(context.Background(), "")
	require.NoError(t, err)
	_, _, err = c.GetUserData(context.Background(), next)
	require.Error(t, err)
}

func TestScenarios(t *testing.T) {
	users, err := DefaultUsers()
	require.NoError(t, err)
	bad, err := ApplyScenario(users, ScenarioBadDate)
	require.NoError(t, err)
	assert.Len(t, bad, len(users)+1)
	empty, err := ApplyScenario(users, ScenarioEmpty)
	require.NoError(t, err)
	assert.Empty(t, empty)
	_, err = ApplyScenario(users, "nope")
	require.Error(t, err)

	e := validEnv(t, func(c *Config) { c.Users = empty })
	got, next, err := e.client(t, "co", "100001").GetUserData(context.Background(), "")
	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Empty(t, next)
}

func TestFixturesMatchExpected(t *testing.T) {
	users, err := DefaultUsers()
	require.NoError(t, err)
	var exp struct {
		Count int                       `json:"count"`
		Users map[string]map[string]any `json:"users"`
	}
	require.NoError(t, json.Unmarshal(ExpectedJSON, &exp))
	assert.Equal(t, exp.Count, len(users))
	assert.Len(t, exp.Users, len(users))
}

func TestLogNeverContainsSecrets(t *testing.T) {
	e := validEnv(t, nil)
	c := e.client(t, "co", "100001")
	_, _, err := c.GetUserData(context.Background(), "")
	require.NoError(t, err)
	logged := e.log.String()
	assert.Contains(t, logged, "POST /oauth/token -> 200")
	assert.NotContains(t, logged, c.SAMLAssertion)
	assert.NotContains(t, logged, "assertion=")
	assert.NotContains(t, logged, "Bearer")
}
