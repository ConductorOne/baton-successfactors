// Package mocksf is a mock of the two SAP SuccessFactors endpoints that
// baton-successfactors calls: POST /oauth/token and GET /odata/v2/EmpJob.
//
// The two error bodies for "SAML assertion in the URL" and "unknown company"
// are copied from the real service. All other error texts are synthetic: we do
// not know how the real service words them.
package mocksf

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	// TokenPath and EmpJobPath are the endpoints the mock serves.
	TokenPath  = "/oauth/token" //nolint:gosec // endpoint, not a credential.
	EmpJobPath = "/odata/v2/EmpJob"

	grantType = "urn:ietf:params:oauth:grant-type:saml2-bearer"
	audience  = "www.successfactors.com"
)

// Config holds the values the mock expects from the connector.
type Config struct {
	Company  string
	ClientID string
	APIKey   string
	Issuer   string
	Subject  string

	// Cert verifies the signature of the SAML assertion.
	Cert *x509.Certificate
	// RejectExpiredCert makes the mock reject a signature made with a cert that
	// is outside its validity dates. The default is to accept it, because the
	// real behavior is unknown.
	RejectExpiredCert bool

	// Users are the raw OData items served by /odata/v2/EmpJob.
	Users    []json.RawMessage
	PageSize int
	// FailPage, with FailStatus, makes the mock return FailStatus for that
	// 1-based page. Zero disables it.
	FailPage   int
	FailStatus int

	// PublicURL is the base URL for __next links and the SAML Recipient check.
	// Empty means: build it from the request.
	PublicURL string

	// Log receives one line per request. Nil means stderr.
	Log io.Writer
	// Now is the clock for the assertion time window. Nil means time.Now.
	Now func() time.Time
}

type server struct {
	cfg    Config
	mu     sync.Mutex
	tokens map[string]struct{}
}

// NewHandler returns the mock handler.
func NewHandler(cfg Config) (http.Handler, error) {
	if cfg.Cert == nil {
		return nil, fmt.Errorf("mocksf: Cert is required")
	}
	if cfg.PageSize <= 0 {
		cfg.PageSize = 5
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Log == nil {
		cfg.Log = os.Stderr
	}
	s := &server{cfg: cfg, tokens: map[string]struct{}{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc(TokenPath, s.handleToken)
	mux.HandleFunc(EmpJobPath, s.handleEmpJob)
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "Resource not found.")
	})
	return s.logRequests(mux), nil
}

// baseURL is the externally visible base URL of the mock.
func (s *server) baseURL(r *http.Request) string {
	if s.cfg.PublicURL != "" {
		return s.cfg.PublicURL
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func (s *server) issueToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	t := hex.EncodeToString(b)
	s.mu.Lock()
	s.tokens[t] = struct{}{}
	s.mu.Unlock()
	return t
}

func (s *server) validToken(t string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.tokens[t]
	return ok
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"errorHttpCode": fmt.Sprint(status),
		"errorMessage":  msg,
	})
}

// writeODataError writes an OData-shaped error body, as the EmpJob endpoint does.
func writeODataError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": map[string]string{"lang": "en-US", "value": msg},
		},
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// logRequests logs the method, path and status. It never logs the query, the
// form or the headers, because they carry the assertion and the token.
func (s *server) logRequests(next http.Handler) http.Handler {
	logger := log.New(s.cfg.Log, "mocksf: ", log.LstdFlags)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		logger.Printf("%s %s -> %d", r.Method, r.URL.Path, sw.status)
	})
}
