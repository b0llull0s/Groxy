package tls

import (
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "crypto/x509/pkix"
    cryptotls "crypto/tls"
    "encoding/pem"
    "fmt"
    "math/big"
    "os"
    "sync"
    "time"
    "context"
)

type Manager struct {
    config       *Config
    currentCert  *cryptotls.Certificate
    certMutex    sync.RWMutex
    rotationDone chan struct{}   
    OnRotation   func(*cryptotls.Certificate)
    OnError      func(error)
    rotateCancel context.CancelFunc
}

func NewManager(config *Config) *Manager {
    return &Manager{
        config:       config,
        rotationDone: make(chan struct{}),
    }
}

func (m *Manager) GenerateCertificate() error {
    cfg := m.config.GetCertificateConfig()
    
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

    certOut, err := os.Create(m.config.CertFile)
    if err != nil {
        return fmt.Errorf("failed to open cert.pem for writing: %v", err)
    }
    defer certOut.Close()

    if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
        return fmt.Errorf("failed to encode certificate: %v", err)
    }

    keyOut, err := os.Create(m.config.KeyFile)
    if err != nil {
        return fmt.Errorf("failed to open key.pem for writing: %v", err)
    }
    defer keyOut.Close()

    if err := pem.Encode(keyOut, &pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
    }); err != nil {
        return fmt.Errorf("failed to encode private key: %v", err)
    }

    return nil
}

func (m *Manager) GetCertificate(*cryptotls.ClientHelloInfo) (*cryptotls.Certificate, error) {
    m.certMutex.RLock()
    defer m.certMutex.RUnlock()
    return m.currentCert, nil
}

func (m *Manager) LoadServerConfig() (*cryptotls.Config, error) {
    cert, err := cryptotls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
    if err != nil {
        return nil, err
    }

    m.certMutex.Lock()
    m.currentCert = &cert
    m.certMutex.Unlock()

    return m.config.LoadServerConfig(m.GetCertificate)
}

func (m *Manager) StartRotation(interval time.Duration) {
    if m.rotateCancel != nil {
        m.rotateCancel()
    }

    ctx, cancel := context.WithCancel(context.Background())
    m.rotateCancel = cancel

    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                if err := m.rotateCertificate(); err != nil && m.OnError != nil {
                    m.OnError(err)
                }
            case <-ctx.Done():
                return
            }
        }
    }()
}

func (m *Manager) StopRotation() {
    if m.rotateCancel != nil {
        m.rotateCancel()
        m.rotateCancel = nil
    }
}

func (m *Manager) rotateCertificate() error {
    if err := m.GenerateCertificate(); err != nil {
        return fmt.Errorf("failed to generate certificate: %v", err)
    }

    cert, err := cryptotls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
    if err != nil {
        return fmt.Errorf("failed to load new certificate: %v", err)
    }

    m.certMutex.Lock()
    m.currentCert = &cert
    m.certMutex.Unlock()

    if m.OnRotation != nil {
        m.OnRotation(&cert)
    }

    return nil
}