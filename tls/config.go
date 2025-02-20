package tls

import (
    cryptotls "crypto/tls" 
)

type Config struct {
    CertFile string
    KeyFile  string
}

func NewConfig(certFile, keyFile string) *Config {
    return &Config{
        CertFile: certFile,
        KeyFile:  keyFile,
    }
}

func (c *Config) LoadServerConfig() (*cryptotls.Config, error) {
    cert, err := cryptotls.LoadX509KeyPair(c.CertFile, c.KeyFile)
    if err != nil {
        return nil, err
    }
    
    return &cryptotls.Config{
        Certificates: []cryptotls.Certificate{cert},
        MinVersion:  cryptotls.VersionTLS12,
    }, nil
}

func (c *Config) LoadClientConfig() *cryptotls.Config {
    return &cryptotls.Config{
        InsecureSkipVerify: true, // Only for development
        MinVersion:         cryptotls.VersionTLS12,
    }
}