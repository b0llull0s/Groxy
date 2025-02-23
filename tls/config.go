package tls

import (
    cryptotls "crypto/tls"
    "fmt"
    "time"
)

type Config struct {
    CertFile string
    KeyFile  string
    CertConfig CertificateConfig
}

type CertificateConfig struct {
    CommonName   string
    Organization []string
    Country      []string
    Validity     time.Duration
    KeySize      int
    IsCA         bool
}

func NewConfig(certFile, keyFile string) *Config {
    return &Config{
        CertFile: certFile,
        KeyFile:  keyFile,
        CertConfig: CertificateConfig{
            CommonName:   "localhost",
            Organization: []string{"YourOrg"},
            Country:     []string{"US"},
            Validity:    time.Hour * 24 * 90, // 90 days
            KeySize:     2048,
            IsCA:        false,
        },
    }
}

func (c *Config) LoadServerConfig(getCertificate func(*cryptotls.ClientHelloInfo) (*cryptotls.Certificate, error)) (*cryptotls.Config, error) {
    cert, err := cryptotls.LoadX509KeyPair(c.CertFile, c.KeyFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load key pair: %v", err)
    }

    return &cryptotls.Config{
        Certificates: []cryptotls.Certificate{cert},
        GetCertificate: getCertificate,
        MinVersion:     cryptotls.VersionTLS12,
    }, nil
}

func (c *Config) LoadClientConfig() *cryptotls.Config {
    return &cryptotls.Config{
        InsecureSkipVerify: true, // Only for development
        MinVersion:         cryptotls.VersionTLS12,
    }
}

func (c *Config) GetCertificateConfig() CertificateConfig {
    return c.CertConfig
}