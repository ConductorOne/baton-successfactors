package mocksf

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

//go:embed testdata/users.json
var defaultUsers []byte

// ExpectedJSON is the expected sync result for the default users.
//
//go:embed testdata/expected.json
var ExpectedJSON []byte

// Scenarios are the dataset names that ApplyScenario accepts.
const (
	ScenarioDefault = "default"
	ScenarioBadDate = "bad-date"
	ScenarioEmpty   = "empty"
)

// DefaultUsers returns the 12 embedded fixture users.
func DefaultUsers() ([]json.RawMessage, error) {
	return parseUsers(defaultUsers)
}

// LoadUsers reads a custom users file. It has the same shape as testdata/users.json.
func LoadUsers(path string) ([]json.RawMessage, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("mocksf: %w", err)
	}
	return parseUsers(b)
}

func parseUsers(b []byte) ([]json.RawMessage, error) {
	var users []json.RawMessage
	if err := json.Unmarshal(b, &users); err != nil {
		return nil, fmt.Errorf("mocksf: users must be a JSON array of EmpJob items: %w", err)
	}
	return users, nil
}

// ApplyScenario changes the dataset for a fault scenario.
func ApplyScenario(users []json.RawMessage, name string) ([]json.RawMessage, error) {
	switch name {
	case "", ScenarioDefault:
		return users, nil
	case ScenarioEmpty:
		return nil, nil
	case ScenarioBadDate:
		bad := json.RawMessage(`{"userId":"99999","userNav":{"username":"baddate","email":"baddate@bestrunsap.example"},"employmentNav":{"endDate":"not-a-date"}}`)
		return append(append([]json.RawMessage{}, users...), bad), nil
	}
	return nil, fmt.Errorf("mocksf: unknown scenario %q (use default, bad-date or empty)", name)
}

// GenerateKeyPair makes a self-signed RSA cert and key in PEM form. Use dates
// in the past to get an expired cert.
func GenerateKeyPair(notBefore, notAfter time.Time) ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "mocksf"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM, nil
}

// ParseCertPEM parses the first certificate in a PEM block.
func ParseCertPEM(b []byte) (*x509.Certificate, error) {
	blk, _ := pem.Decode(b)
	if blk == nil {
		return nil, fmt.Errorf("mocksf: no PEM block in certificate")
	}
	return x509.ParseCertificate(blk.Bytes)
}
