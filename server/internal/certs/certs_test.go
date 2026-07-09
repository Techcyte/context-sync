package certs

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateProducesVerifiableChain(t *testing.T) {
	dir := t.TempDir()

	caPath, err := Generate(dir)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if caPath != filepath.Join(dir, CACertFile) {
		t.Fatalf("unexpected caPath: %s", caPath)
	}

	// The TLS pair the server actually serves must load.
	pair, err := tls.LoadX509KeyPair(filepath.Join(dir, ServerCertFile), filepath.Join(dir, ServerKeyFile))
	if err != nil {
		t.Fatalf("LoadX509KeyPair: %v", err)
	}
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		t.Fatalf("parse leaf: %v", err)
	}

	// The leaf must chain up to (and only to) our generated CA.
	caPEM, err := os.ReadFile(caPath)
	if err != nil {
		t.Fatalf("read ca: %v", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		t.Fatalf("could not load CA into pool")
	}
	if _, err := leaf.Verify(x509.VerifyOptions{DNSName: "localhost", Roots: roots}); err != nil {
		t.Fatalf("leaf did not verify for localhost: %v", err)
	}

	// 127.0.0.1 must also be covered so wss://127.0.0.1 works.
	if err := leaf.VerifyHostname("127.0.0.1"); err != nil {
		t.Fatalf("leaf did not cover 127.0.0.1: %v", err)
	}
}
