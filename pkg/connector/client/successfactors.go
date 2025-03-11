package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/beevik/etree"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/crewjam/saml"
	"github.com/google/uuid"
	dsig "github.com/russellhaering/goxmldsig"
)

const (
	APIPath     = ""
	AuditorRole = "AUDITOR"
)

type SuccessFactorsClient struct {
	baseURL       string
	client        *uhttp.BaseHttpClient
	compID        string
	clientID      string
	pubKey        string
	privKey       string
	issuerURL     string
	subNID        string
	SAMLAPIKey    string
	SAMLAssertion string
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
	signedAssertion, err := createAndSignSAMLAssertion(issuerURL, "www.successfactors.com", baseURL+"/oauth/token", subNID, samlAPIKey, privKey, pubKey)
	if err != nil {
		return nil, err
	}
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, nil))
	if err != nil {
		return nil, err
	}

	client, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}
	return &SuccessFactorsClient{
		baseURL:       baseURL,
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
func (c *SuccessFactorsClient) doRequest(ctx context.Context, method string, u *url.URL, reqOpts []uhttp.RequestOption, body interface{}, response interface{}) error {
	if body != nil {
		reqOpts = append(reqOpts, uhttp.WithJSONBody(body), uhttp.WithContentTypeJSONHeader())
	}
	req, err := c.client.NewRequest(ctx, method, u, reqOpts...)
	if err != nil {
		return err
	}
	doOpts := []uhttp.DoOption{}
	if response != nil {
		doOpts = append(doOpts, uhttp.WithJSONResponse(response))
	}

	resp, err := c.client.Do(req, doOpts...)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}

func createAndSignSAMLAssertion(issuer, audience, recipient, subjectNameId, apiKey, privKey, certificate string) (string, error) {
	// Generate timestamps
	now := time.Now().UTC().Add(-5 * time.Second)
	notBefore := now
	notOnOrAfter := now.Add((24 * time.Hour) - (5 * time.Second))

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
	if err != nil {
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

func (c *SuccessFactorsClient) GetBearer(ctx context.Context) (string, error) {
	var response Bearer
	reqOpts := []uhttp.RequestOption{
		uhttp.WithContentTypeFormHeader(),
	}
	u, err := url.Parse(c.baseURL + "/oauth/token")
	if err != nil {
		return "", fmt.Errorf("failed to get bearer: %w", err)
	}
	values := u.Query()
	values.Add("company_id", c.compID)
	values.Add("client_id", c.clientID)
	values.Add("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
	values.Add("assertion", c.SAMLAssertion)
	u.RawQuery = values.Encode()
	err = c.doRequest(ctx, http.MethodPost, u, reqOpts, nil, &response)
	if err != nil {
		return "", fmt.Errorf("failed to get bearer: %w", err)
	}
	return response.AccessToken, nil
}

func (c *SuccessFactorsClient) GetUserData(ctx context.Context) ([]Results, error) {
	var response SuccessFactorsUserIdList
	var responses []Results
	bearer, err := c.GetBearer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bearer: %w", err)
	}
	reqOpts := []uhttp.RequestOption{
		uhttp.WithHeader("Authorization", "Bearer "+bearer),
	}
	u, err := url.Parse(c.baseURL + "/odata/v2/EmpJob")
	if err != nil {
		return nil, fmt.Errorf("failed to parse initial string: %w", err)
	}
	values := u.Query()
	values.Add("$expand", "userNav,employmentNav,companyNav,businessUnitNav,divisionNav,departmentNav,locationNav,costCenterNav,positionNav")
	values.Add("$expand", "employeeClassNav,emplStatusNav/picklistLabels,managerUserNav,companyNav,employmentNav,companyNav/countryNav,employeeClassNav/picklistLabels")
	values.Add("$format", "json")
	values.Add("$select", "userId,userNav/firstName,userNav/lastName,userNav/mi,userNav/username,userNav/email,employmentNav/startDate,jobTitle,localJobTitle,companyNav/name_localized")
	values.Add("$select", "businessUnitNav/name,divisionNav/name,departmentNav/name,locationNav/name,costCenterNav/name_defaultValue,positionNav/code,positionNav/externalName_defaultValue")
	values.Add("$select", "employeeClassNav/picklistLabels/label,emplStatusNav/picklistLabels/label,managerUserNav/userId,managerUserNav/email,companyNav/countryNav/territoryName")
	values.Add("$select", "employmentNav/endDate,userNav/custom07")
	u.RawQuery = values.Encode()
	err = c.doRequest(ctx, http.MethodGet, u, reqOpts, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	for response.Ds.Next != "" {
		responses = slices.Concat(responses, response.Ds.Results)
		u, err = url.Parse(response.Ds.Next)
		if err != nil {
			return nil, fmt.Errorf("failed to parse next token: %w", err)
		}
		response = SuccessFactorsUserIdList{}
		err = c.doRequest(ctx, http.MethodGet, u, reqOpts, nil, &response)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
	}

	return responses, nil
}
