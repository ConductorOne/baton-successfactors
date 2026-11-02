package mocksf

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
)

func (s *server) handleToken(w http.ResponseWriter, r *http.Request) {
	// Copied from the real service.
	if r.URL.RawQuery != "" {
		writeError(w, http.StatusBadRequest, "SAML assertion is not allowed in URL parameter.")
		return
	}
	if r.Method != http.MethodPost || !strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		writeError(w, http.StatusBadRequest, "Synthetic: POST with an application/x-www-form-urlencoded body is required.")
		return
	}
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "Synthetic: the form body cannot be parsed.")
		return
	}
	form := r.PostForm

	// The text of this message is copied from the real service.
	if form.Get("company_id") != s.cfg.Company {
		writeError(w, http.StatusUnauthorized, fmt.Sprintf("Company %s does not exist in this data center. Please verify the company id and the endpoint URL you are using.", form.Get("company_id")))
		return
	}
	if form.Get("grant_type") != grantType {
		writeError(w, http.StatusUnauthorized, "Synthetic: unsupported grant_type.")
		return
	}
	if form.Get("client_id") != s.cfg.ClientID {
		writeError(w, http.StatusUnauthorized, "Synthetic: unknown client_id.")
		return
	}
	if err := s.verifyAssertion(form.Get("assertion"), s.baseURL(r)+TokenPath); err != nil {
		writeError(w, http.StatusUnauthorized, "Synthetic: invalid SAML assertion: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": s.issueToken(),
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

// verifyAssertion checks the signature and the claims of the SAML assertion.
// The error texts do not contain the assertion.
func (s *server) verifyAssertion(encoded, recipient string) error {
	if encoded == "" {
		return fmt.Errorf("assertion is missing")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("assertion is not base64")
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(raw); err != nil || doc.Root() == nil {
		return fmt.Errorf("assertion is not XML")
	}

	vctx := dsig.NewDefaultValidationContext(&dsig.MemoryX509CertificateStore{Roots: []*x509.Certificate{s.cfg.Cert}})
	if !s.cfg.RejectExpiredCert {
		// Pin the clock inside the cert validity, so that an expired cert is accepted.
		mid := s.cfg.Cert.NotBefore.Add(s.cfg.Cert.NotAfter.Sub(s.cfg.Cert.NotBefore) / 2)
		vctx.Clock = dsig.NewFakeClockAt(mid)
	}
	el, err := vctx.Validate(doc.Root())
	if err != nil {
		return fmt.Errorf("signature check failed: %v", err)
	}

	// Read the claims from the validated element only.
	if got := text(el, "Issuer"); got != s.cfg.Issuer {
		return fmt.Errorf("issuer does not match")
	}
	if got := text(el, "Subject", "NameID"); got != s.cfg.Subject {
		return fmt.Errorf("subject does not match")
	}
	if got := text(el, "Conditions", "AudienceRestriction", "Audience"); got != audience {
		return fmt.Errorf("audience does not match")
	}
	scd := child(child(child(el, "Subject"), "SubjectConfirmation"), "SubjectConfirmationData")
	if scd == nil || scd.SelectAttrValue("Recipient", "") != recipient {
		return fmt.Errorf("recipient does not match")
	}
	if got := apiKey(el); got != s.cfg.APIKey {
		return fmt.Errorf("api_key does not match")
	}
	cond := child(el, "Conditions")
	if cond == nil {
		return fmt.Errorf("conditions are missing")
	}
	now := s.cfg.Now()
	nb, err1 := time.Parse(time.RFC3339Nano, cond.SelectAttrValue("NotBefore", ""))
	na, err2 := time.Parse(time.RFC3339Nano, cond.SelectAttrValue("NotOnOrAfter", ""))
	if err1 != nil || err2 != nil {
		return fmt.Errorf("conditions have no valid time window")
	}
	if now.Before(nb) || !now.Before(na) {
		return fmt.Errorf("assertion is outside its time window")
	}
	return nil
}

// child returns the first child element with the given local name.
func child(el *etree.Element, tag string) *etree.Element {
	if el == nil {
		return nil
	}
	for _, c := range el.ChildElements() {
		if c.Tag == tag {
			return c
		}
	}
	return nil
}

func text(el *etree.Element, path ...string) string {
	for _, p := range path {
		el = child(el, p)
	}
	if el == nil {
		return ""
	}
	return strings.TrimSpace(el.Text())
}

func apiKey(el *etree.Element) string {
	stmt := child(el, "AttributeStatement")
	if stmt == nil {
		return ""
	}
	for _, a := range stmt.ChildElements() {
		if a.Tag == "Attribute" && a.SelectAttrValue("Name", "") == "api_key" {
			return text(a, "AttributeValue")
		}
	}
	return ""
}
