// Command mock-successfactors runs a mock SAP SuccessFactors server for CI and
// local tests of baton-successfactors. See docs/mock-server.md.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/conductorone/baton-successfactors/internal/mocksf"
)

func main() {
	var (
		addr       = flag.String("addr", "127.0.0.1:8089", "listen address")
		certPath   = flag.String("cert", "", "PEM cert that verifies the SAML assertion (required)")
		company    = flag.String("company", "MOCKCO", "expected company_id")
		clientID   = flag.String("client-id", "mock-client", "expected client_id")
		apiKey     = flag.String("api-key", "mock-api-key", "expected api_key attribute")
		issuer     = flag.String("issuer", "https://mock-issuer.example", "expected SAML issuer")
		subject    = flag.String("subject", "100001", "expected SAML subject name ID")
		fixtures   = flag.String("fixtures", "", "path to a custom users JSON file (default: embedded users)")
		scenario   = flag.String("scenario", "default", "dataset: default, bad-date or empty")
		pageSize   = flag.Int("page-size", 5, "users per page")
		failPage   = flag.Int("fail-page", 0, "1-based page that returns -fail-status (0 = off)")
		failStatus = flag.Int("fail-status", 500, "status for -fail-page")
		rejectExp  = flag.Bool("reject-expired-cert", false, "reject a signature made with an expired cert")
		publicURL  = flag.String("public-url", "", "base URL for __next links (default: from the request)")
	)
	flag.Parse()

	if err := run(*addr, *certPath, *fixtures, *scenario, *publicURL, mocksf.Config{
		Company:           *company,
		ClientID:          *clientID,
		APIKey:            *apiKey,
		Issuer:            *issuer,
		Subject:           *subject,
		PageSize:          *pageSize,
		FailPage:          *failPage,
		FailStatus:        *failStatus,
		RejectExpiredCert: *rejectExp,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "mock-successfactors:", err)
		os.Exit(1)
	}
}

func run(addr, certPath, fixtures, scenario, publicURL string, cfg mocksf.Config) error {
	if certPath == "" {
		return fmt.Errorf("-cert is required")
	}
	pemBytes, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	if cfg.Cert, err = mocksf.ParseCertPEM(pemBytes); err != nil {
		return err
	}
	if fixtures != "" {
		cfg.Users, err = mocksf.LoadUsers(fixtures)
	} else {
		cfg.Users, err = mocksf.DefaultUsers()
	}
	if err != nil {
		return err
	}
	if cfg.Users, err = mocksf.ApplyScenario(cfg.Users, scenario); err != nil {
		return err
	}
	cfg.PublicURL = publicURL

	h, err := mocksf.NewHandler(cfg)
	if err != nil {
		return err
	}
	log.Printf("mock-successfactors: %d users, listening on %s", len(cfg.Users), addr)
	srv := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 10 * time.Second}
	return srv.ListenAndServe()
}
