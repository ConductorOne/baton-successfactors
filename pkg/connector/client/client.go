package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/beevik/etree"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/crewjam/saml"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	dsig "github.com/russellhaering/goxmldsig"
)

type SuccessFactorsClient struct {
	compID        string
	clientID      string
	pubKey        string
	privKey       string
	issuerURL     string
	subNID        string
	SAMLAPIKey    string
	SAMLAssertion string

	baseURL     *url.URL
	client      *uhttp.BaseHttpClient
	bearerToken string
}

func New(
	ctx context.Context,
	baseURL string,
	compID string,
	clientID string,
	pubKey string,
	privKey string,
	issuerURL string,
	subNID string,
	samlAPIKey string,
) (*SuccessFactorsClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing instance-url")
	}
	samlbase := base.JoinPath(base.RawPath, "/oauth/token")
	signedAssertion, err := createAndSignSAMLAssertion(issuerURL, "www.successfactors.com", samlbase.String(), subNID, samlAPIKey, privKey, pubKey)
	if err != nil {
		return nil, err
	}
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	client, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}
	return &SuccessFactorsClient{
		baseURL:       base,
		client:        client,
		compID:        compID,
		clientID:      clientID,
		pubKey:        pubKey,
		privKey:       privKey,
		issuerURL:     issuerURL,
		subNID:        subNID,
		SAMLAPIKey:    samlAPIKey,
		SAMLAssertion: signedAssertion,
	}, nil
}

func createAndSignSAMLAssertion(issuer, audience, recipient, subjectNameId, apiKey, privKey, certificate string) (string, error) {
	// Generate timestamps.
	// We subtract 5 from now to account for clock skew between the client and server
	now := time.Now().UTC().Add(-5 * time.Second)
	notBefore := now
	notOnOrAfter := now.Add((24 * time.Hour))

	// Create assertion
	assertion := &saml.Assertion{
		ID:           fmt.Sprintf("_%s", uuid.New().String()),
		IssueInstant: now,
		Version:      "2.0",
		Issuer: saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  issuer,
		},
		Subject: &saml.Subject{
			NameID: &saml.NameID{
				Format: "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
				Value:  subjectNameId,
			},
			SubjectConfirmations: []saml.SubjectConfirmation{
				{
					Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
					SubjectConfirmationData: &saml.SubjectConfirmationData{
						NotOnOrAfter: notOnOrAfter,
						Recipient:    recipient,
					},
				},
			},
		},
		Conditions: &saml.Conditions{
			NotBefore:    notBefore,
			NotOnOrAfter: notOnOrAfter,
			AudienceRestrictions: []saml.AudienceRestriction{
				{
					Audience: saml.Audience{Value: audience},
				},
			},
		},
		AttributeStatements: []saml.AttributeStatement{
			{
				Attributes: []saml.Attribute{
					{
						Name:   "api_key",
						Values: []saml.AttributeValue{{Type: "xs:string", Value: apiKey}},
					},
				},
			},
		},
	}

	// Load certificate and private key
	keyPair, err := tls.X509KeyPair([]byte(certificate), []byte(privKey))
	if err != nil {
		return "", fmt.Errorf("failed to load key pair: %w", err)
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Create XML document for signing
	doc := etree.NewDocument()
	doc.SetRoot(assertion.Element())

	// Create signing context
	keyStore := dsig.TLSCertKeyStore(keyPair)
	signingContext := dsig.NewDefaultSigningContext(keyStore)
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")

	// Sign the assertion
	signedElement, err := signingContext.SignEnveloped(doc.Root())
	if err != nil || len(signedElement.ChildElements()) == 0 {
		return "", fmt.Errorf("failed to sign assertion: %w", err)
	}

	// Get the signature element and add it to the assertion
	sigEl := signedElement.ChildElements()[len(signedElement.ChildElements())-1]
	assertion.Signature = sigEl

	// Convert to XML string
	doc = etree.NewDocument()
	doc.SetRoot(signedElement)
	signedXML, err := doc.WriteToBytes()
	if err != nil {
		return "", fmt.Errorf("failed to marshal signed assertion: %w", err)
	}

	// Base64 encode the signed assertion
	encoded := base64.StdEncoding.EncodeToString(signedXML)
	return encoded, nil
}

func (c *SuccessFactorsClient) doRequest(
	ctx context.Context,
	method string,
	endpointUrl string,
	res interface{},
	body interface{},
	rateLimitDescription *v2.RateLimitDescription,
) error {
	var (
		resp        *http.Response
		err         error
		errResponse ErrorResponse
	)

	urlAddress, err := url.Parse(endpointUrl)
	if err != nil {
		return err
	}

	// TODO: Will change this later, it's only for tests (JAV)
	var opts []uhttp.RequestOption
	if method == "POST" {
		opts = append(opts,
			uhttp.WithContentTypeFormHeader(),
		)
	} else {
		opts = append(opts,
			uhttp.WithBearerToken(c.bearerToken), uhttp.WithContentType("application/scim+json"),
		)
	}

	if body != nil {
		opts = append(opts, uhttp.WithJSONBody(body))
	}

	req, err := c.client.NewRequest(
		ctx,
		method,
		urlAddress,
		opts...,
	)
	if err != nil {
		return err
	}

	doOptions := []uhttp.DoOption{uhttp.WithErrorResponse(&errResponse)}
	if rateLimitDescription != nil {
		doOptions = append(doOptions, uhttp.WithRatelimitData(rateLimitDescription))
	}
	if res != nil {
		doOptions = append(doOptions, uhttp.WithResponse(&res))
	}

	resp, err = c.client.Do(req, doOptions...)
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	return nil
}

func (c *SuccessFactorsClient) getBearer(ctx context.Context) (*BearerToken, *v2.RateLimitDescription, error) {
	var response *BearerToken

	requestURL := c.baseURL.JoinPath(c.baseURL.RawPath, "/oauth/token")
	values := requestURL.Query()
	values.Add("company_id", c.compID)
	values.Add("client_id", c.clientID)
	values.Add("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
	values.Add("assertion", c.SAMLAssertion)
	requestURL.RawQuery = values.Encode()

	rateLimitDescription := &v2.RateLimitDescription{}
	err := c.doRequest(
		ctx,
		http.MethodPost,
		requestURL.String(),
		&response,
		nil,
		rateLimitDescription,
	)
	if err != nil {
		return nil, rateLimitDescription, fmt.Errorf("failed to get bearer: %w", err)
	}

	return response, rateLimitDescription, nil
}

func (c *SuccessFactorsClient) GetUsers(ctx context.Context, startIndex string) ([]*User, string, *v2.RateLimitDescription, error) {
	var response Response

	// Request a new Bearer Token. TODO: This should be done only when expired.
	bearer, rateLimitData, err := c.getBearer(ctx)
	if err != nil {
		return nil, "", rateLimitData, fmt.Errorf("failed to get new bearer token: %w", err)
	}
	c.bearerToken = bearer.AccessToken

	requestURL := c.baseURL.JoinPath(c.baseURL.RawPath, "/rest/iam/scim/v2/Users")
	values := requestURL.Query()
	values.Add("count", maxPageSize)
	if startIndex != "" {
		values.Add("startIndex", startIndex)
	}
	requestURL.RawQuery = values.Encode()

	rateLimitDescription := &v2.RateLimitDescription{}
	err = c.doRequest(
		ctx,
		http.MethodGet,
		requestURL.String(),
		&response,
		nil,
		rateLimitDescription,
	)
	if err != nil {
		return nil, "", rateLimitDescription, fmt.Errorf("failed to request users: %w", err)
	}

	var users []*User
	err = json.Unmarshal(response.Resources, &users)
	if err != nil {
		return nil, "", rateLimitDescription, fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	nextPageIndex := calculateNextPageStartIndex(&response)

	return users, nextPageIndex, rateLimitDescription, nil
}

func (c *SuccessFactorsClient) GetGroups(ctx context.Context, startIndex string) ([]*Group, string, *v2.RateLimitDescription, error) {
	var response Response

	// Request a new Bearer Token. TODO: This should be done only when expired.
	bearer, rateLimitData, err := c.getBearer(ctx)
	if err != nil {
		return nil, "", rateLimitData, fmt.Errorf("failed to get new bearer token: %w", err)
	}
	c.bearerToken = bearer.AccessToken

	requestURL := c.baseURL.JoinPath(c.baseURL.RawPath, "/rest/iam/scim/v2/Groups")
	values := requestURL.Query()
	values.Add("count", maxPageSize)
	if startIndex != "" {
		values.Add("startIndex", startIndex)
	}
	requestURL.RawQuery = values.Encode()

	rateLimitDescription := &v2.RateLimitDescription{}
	err = c.doRequest(
		ctx,
		http.MethodGet,
		requestURL.String(),
		&response,
		nil,
		rateLimitDescription,
	)
	if err != nil {
		return nil, "", rateLimitDescription, fmt.Errorf("failed to request groups: %w", err)
	}

	var groups []*Group
	err = json.Unmarshal(response.Resources, &groups)
	if err != nil {
		return nil, "", rateLimitDescription, fmt.Errorf("failed to unmarshal groups response: %w", err)
	}

	nextPageIndex := calculateNextPageStartIndex(&response)

	return groups, nextPageIndex, rateLimitDescription, nil
}

func (c *SuccessFactorsClient) GetGroup(ctx context.Context, groupID string) (*Group, *v2.RateLimitDescription, error) {
	var response *Group

	// Request a new Bearer Token. TODO: This should be done only when expired.
	bearer, rateLimitData, err := c.getBearer(ctx)
	if err != nil {
		return nil, rateLimitData, fmt.Errorf("failed to get new bearer token: %w", err)
	}
	c.bearerToken = bearer.AccessToken

	requestURL := c.baseURL.JoinPath(c.baseURL.RawPath, "/rest/iam/scim/v2/Groups", groupID)

	rateLimitDescription := &v2.RateLimitDescription{}
	err = c.doRequest(
		ctx,
		http.MethodGet,
		requestURL.String(),
		&response,
		nil,
		rateLimitDescription,
	)
	if err != nil {
		return nil, rateLimitDescription, fmt.Errorf("failed to request group {%s}: %w", groupID, err)
	}

	return response, rateLimitDescription, nil
}
