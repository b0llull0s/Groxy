package tls

import (
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "crypto/x509/pkix"
 	cryptotls "crypto/tls"
    "encoding/pem"
    "math/big"
    "sync"
    "time"
    "fmt"
    "os"
)

// CertificateConfig holds the configuration for certificate generation
type CertificateConfig struct {
    CommonName       string
    Organization     []string
    Country          []string
    Validity         time.Duration
    KeySize          int
    IsCA             bool
}

// Manager handles certificate operations and rotation
type Manager struct {
    config         *Config
    currentCert    *cryptotls.Certificate
    certMutex      sync.RWMutex
    rotationDone   chan struct{}   
    OnRotation     func(*cryptotls.Certificate)
    OnError        func(error)
}

// NewManager creates a new TLS manager
func NewManager(config *Config) *Manager {
    return &Manager{
        config:       config,
        rotationDone: make(chan struct{}),
    }
}

// GenerateCertificate creates a new self-signed certificate
func (m *Manager) GenerateCertificate(cfg CertificateConfig) error {
    privateKey, err := rsa.GenerateKey(rand.Reader, cfg.KeySize)
    if err != nil {
        return fmt.Errorf("failed to generate private key: %v", err)
    }

    template := x509.Certificate{
        SerialNumber: big.NewInt(time.Now().Unix()),
        Subject: pkix.Name{
            CommonName:   cfg.CommonName,
            Organization: cfg.Organization,
            Country:      cfg.Country,
        },
        NotBefore:             time.Now(),
        NotAfter:              time.Now().Add(cfg.Validity),
        KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
        ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
        BasicConstraintsValid: true,
        IsCA:                  cfg.IsCA,
    }

    certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
    if err != nil {
        return fmt.Errorf("failed to create certificate: %v", err)
    }

    // Save the certificate and private key
    certFile := m.config.CertFile
    keyFile := m.config.KeyFile

    certOut, err := os.Create(certFile)
    if err != nil {
        return fmt.Errorf("failed to open cert.pem for writing: %v", err)
    }
    pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
    certOut.Close()

    keyOut, err := os.Create(keyFile)
    if err != nil {
        return fmt.Errorf("failed to open key.pem for writing: %v", err)
    }
    pem.Encode(keyOut, &pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
    })
    keyOut.Close()

    return nil
}

// StartRotation begins certificate rotation with the specified interval
func (m *Manager) StartRotation(interval time.Duration) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                err := m.rotateCertificate()
                if err != nil && m.OnError != nil {
                    m.OnError(err)
                }
            case <-m.rotationDone:
                return
            }
        }
    }()
}

// StopRotation stops the certificate rotation
func (m *Manager) StopRotation() {
    close(m.rotationDone)
}

// rotateCertificate generates and loads a new certificate
func (m *Manager) rotateCertificate() error {
    // Generate new certificate with same config
    cfg := CertificateConfig{
        CommonName:    "localhost", // Could be configurable
        Organization:  []string{"YourOrg"},
        Country:      []string{"US"},
        Validity:     time.Hour * 24 * 90, // 90 days
        KeySize:      2048,
        IsCA:         false,
    }

    if err := m.GenerateCertificate(cfg); err != nil {
        return err
    }

    // Load the new certificate
    cert, err := cryptotls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
    if err != nil {
        return err
    }

    // Update the certificate
    m.certMutex.Lock()
    m.currentCert = &cert
    m.certMutex.Unlock()

    // Notify listeners
    if m.OnRotation != nil {
        m.OnRotation(&cert)
    }

    return nil
}

// GetCertificate implements the GetCertificate function for tls.Config
func (m *Manager) GetCertificate(*cryptotls.ClientHelloInfo) (*cryptotls.Certificate, error) {
    m.certMutex.RLock()
    defer m.certMutex.RUnlock()
    return m.currentCert, nil
}

// LoadServerConfig returns a crypto/tls.Config with dynamic certificate loading
func (m *Manager) LoadServerConfig() (*cryptotls.Config, error) {
    // Load initial certificate
    cert, err := cryptotls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
    if err != nil {
        return nil, err
    }

    m.certMutex.Lock()
    m.currentCert = &cert
    m.certMutex.Unlock()

    return &cryptotls.Config{
        GetCertificate: m.GetCertificate,
        MinVersion:     cryptotls.VersionTLS12,
    }, nil
}
