package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	CACertFile     = "ca.crt"
	CAKeyFile      = "ca.key"
	ServerCertFile = "server.crt"
	ServerKeyFile  = "server.key"
)

// validity is how long the generated certificates are valid for.
const validity = 2 * 365 * 24 * time.Hour // ~2 years

// Ensure makes sure a usable server certificate and key exist in dir. If they
// are missing it generates a fresh CA + server certificate and attempts to
// install the CA into the system trust store.
//
// It returns whether new certificates were generated so the caller can decide
// what to tell the user. A generation failure is fatal (the server cannot serve
// TLS without a certificate), a trust-installation failure is returned as a
// separate, non-fatal warning so the server can still start.
func Ensure(dir string) (generated bool, warning error, err error) {
	if CertificatesExist(dir) {
		return false, nil, nil
	}

	caPath, err := Generate(dir)
	if err != nil {
		return false, nil, err
	}

	if installErr := InstallCA(caPath); installErr != nil {
		return true, installErr, nil
	}

	return true, nil, nil
}

// CertificatesExist reports whether the server certificate and key already exist
// in dir.
func CertificatesExist(dir string) bool {
	return fileExists(filepath.Join(dir, ServerCertFile)) &&
		fileExists(filepath.Join(dir, ServerKeyFile))
}

// Generate creates a local CA and a "localhost" server certificate signed by it,
// writing ca.crt, ca.key, server.crt and server.key into dir. It returns the
// path to the CA certificate, which is the file that needs to be trusted.
func Generate(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating certificate directory: %w", err)
	}

	now := time.Now()

	// --- Certificate Authority -------------------------------------------
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("generating CA key: %w", err)
	}

	caSerial, err := randomSerial()
	if err != nil {
		return "", err
	}

	caTemplate := &x509.Certificate{
		SerialNumber: caSerial,
		Subject: pkix.Name{
			CommonName:   "Techcyte Local CA",
			Organization: []string{"Techcyte"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(validity),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return "", fmt.Errorf("creating CA certificate: %w", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return "", fmt.Errorf("parsing CA certificate: %w", err)
	}

	// --- Leaf (server) certificate ---------------------------------------
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("generating server key: %w", err)
	}

	serverSerial, err := randomSerial()
	if err != nil {
		return "", err
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: serverSerial,
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"Techcyte"},
		},
		NotBefore:   now,
		NotAfter:    now.Add(validity),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		// Subject Alternative Names. Browsers match the host against these, so
		// they must cover every name the client uses to reach the server.
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return "", fmt.Errorf("creating server certificate: %w", err)
	}

	// --- Write everything to disk ----------------------------------------
	caPath := filepath.Join(dir, CACertFile)
	if err := writeCertPEM(caPath, caDER); err != nil {
		return "", err
	}
	if err := writeKeyPEM(filepath.Join(dir, CAKeyFile), caKey); err != nil {
		return "", err
	}
	if err := writeCertPEM(filepath.Join(dir, ServerCertFile), serverDER); err != nil {
		return "", err
	}
	if err := writeKeyPEM(filepath.Join(dir, ServerKeyFile), serverKey); err != nil {
		return "", err
	}

	return caPath, nil
}

// InstallCA adds the CA certificate at caPath to the current user's trust store
// so browsers accept certificates it signed. It shells out to the tools that
// ship with each OS (certutil on Windows, security on macOS) so no extra
// software is required.
func InstallCA(caPath string) error {
	switch runtime.GOOS {
	case "windows":
		return installCAWindows(caPath)
	case "darwin":
		return installCADarwin(caPath)
	default:
		return fmt.Errorf("automatic trust installation is not supported on %s; trust %s manually", runtime.GOOS, caPath)
	}
}

// installCAWindows adds the CA to the current user's Root store with certutil.
// The first time this runs Windows shows a confirmation dialog the user must
// accept. No administrator rights are needed for the per-user store.
func installCAWindows(caPath string) error {
	out, err := exec.Command("certutil", "-user", "-addstore", "Root", caPath).CombinedOutput()
	if err == nil || strings.Contains(string(out), "already in store") {
		return nil
	}
	return fmt.Errorf("certutil failed to install the CA: %v: %s", err, strings.TrimSpace(string(out)))
}

// installCADarwin adds the CA to the user's login keychain and opens Keychain
// Access so the user can mark it as trusted. macOS does not let a non-admin
// program flip the trust setting automatically, so this leaves that final click
// to the user (see the README).
func installCADarwin(caPath string) error {
	out, err := exec.Command("security", "add-certificates", caPath).CombinedOutput()
	if err != nil && !strings.Contains(string(out), "already exists") {
		return fmt.Errorf("security failed to add the CA: %v: %s", err, strings.TrimSpace(string(out)))
	}
	// Best-effort: open Keychain Access so the user can set "Always Trust".
	_ = exec.Command("open", "-a", "Keychain Access").Start()
	return nil
}

func randomSerial() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return nil, fmt.Errorf("generating serial number: %w", err)
	}
	return serial, nil
}

func writeCertPEM(path string, der []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func writeKeyPEM(path string, key *ecdsa.PrivateKey) error {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshaling private key: %w", err)
	}
	// Private keys are written with restrictive permissions.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
