package client

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/square/go-jose/v3"
	"github.com/square/go-jose/v3/jwt"
	"go.uber.org/zap"
)

const (
	defaultPageSize = 100
)

type Client struct {
	httpClient      *uhttp.BaseHttpClient
	baseURL         string
	clientID        string
	companyID       string
	certPrivateKey  string
	samlAssertionURL string
	idpIssuerURL    string
	idpSubjectNameID string

	accessToken     string
	accessTokenExpiry time.Time
}

func New(
	ctx context.Context,
	baseURL string,
	clientID string,
	companyID string,
	certPrivateKey string,
	samlAssertionURL string,
	idpIssuerURL string,
	idpSubjectNameID string,
) (*Client, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, zap.L().Named("successfactors")))
	if err != nil {
		return nil, err
	}

	if samlAssertionURL == "" {
		samlAssertionURL = baseURL + "/oauth/token"
	}

	return &Client{
		httpClient:      uhttp.NewBaseHttpClient(httpClient),
		baseURL:         baseURL,
		clientID:        clientID,
		companyID:       companyID,
		certPrivateKey:  certPrivateKey,
		samlAssertionURL: samlAssertionURL,
		idpIssuerURL:    idpIssuerURL,
		idpSubjectNameID: idpSubjectNameID,
	}, nil
}

// GetUsers retrieves users from the SuccessFactors API.
func (c *Client) GetUsers(ctx context.Context, queryParams map[string]interface{}) ([]interface{}, *v2.RateLimitDescription, error) {
	var users []interface{} // Replace interface{} with your actual User type

	// Build the URL for the API request
	uri, err := url.Parse(c.baseURL + "/odata/v2/User") // Replace "/odata/v2/User" with the correct endpoint
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	q := uri.Query()
	for key, value := range queryParams {
		q.Add(key, fmt.Sprintf("%v", value))
	}
	uri.RawQuery = q.Encode()

	// Build the request
	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, uri, uhttp.WithHeader("Authorization", fmt.Sprintf("Bearer %s", c.accessToken)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	var rl v2.RateLimitDescription
	resp, err := c.httpClient.Do(req, uhttp.WithRatelimitData(&rl))
	if err != nil {
		return nil, &rl, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Process the response (replace with your actual logic)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &rl, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the JSON response into your User struct
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, &rl, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return users, &rl, nil
}

func (c *Client) getAccessToken(ctx context.Context) error {
	if c.accessToken != "" && time.Now().Before(c.accessTokenExpiry) {
		// Token is valid, reuse it
		return nil
	}

	samlAssertion, err := c.generateSAMLAssertion(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate SAML assertion: %w", err)
	}

	data := url.Values{}
	data.Set("company_id", c.companyID)
	data.Set("client_id", c.clientID)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
	data.Set("assertion", samlAssertion)

	uri, err := url.Parse(c.samlAssertionURL)
	if err != nil {
		return fmt.Errorf("failed to parse token URL: %w", err)
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodPost, uri, uhttp.WithFormBody(data.Encode()), uhttp.WithHeader("Content-Type", "application/x-www-form-urlencoded"))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response body: %w", err)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	c.accessToken = tokenResponse.AccessToken
	c.accessTokenExpiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	return nil
}

func (c *Client) generateSAMLAssertion(ctx context.Context) (string, error) {
	now := time.Now()
	claims := jwt.Claims{
		Issuer:    c.idpIssuerURL,
		Subject:   c.idpSubjectNameID, // TODO: Verify the requirements for the Subject.  Is this the baton user?
		Audience:  []string{"www.successfactors.com"},        // Is this correct or configurable?
		Expiry:    now.Add(5 * time.Minute).Unix(),             // 5 minutes validity
		NotBefore: now.Add(-5 * time.Second).Unix(),            // Allow for clock skew
		IssuedAt:  now.Unix(),
	}

	signer, err := c.newSigner()
	if err != nil {
		return "", err
	}

	builder := jwt.Signed(signer).Claims(claims)
	raw, err := builder.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize jwt: %w", err)
	}

	return raw, nil
}

func (c *Client) newSigner() (jose.Signer, error) {
	block, _ := pem.Decode([]byte(c.certPrivateKey))
	if block == nil {
		return nil, errors.New("pem decode failed")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key failed: %w", err)
	}

	//	opts := &rsa.Options{ Hash: crypto.SHA256 }
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return nil, fmt.Errorf("new signer failed: %w", err)
	}

	return sig, nil
}