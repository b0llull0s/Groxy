package tls

import (
    "crypto/tls"
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

func (c *Config) LoadServerConfig() (*tls.Config, error) {
    cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
    if err != nil {
        return nil, err
    }
    
    return &tls.Config{
        Certificates: []tls.Certificate{cert},
        MinVersion:  tls.VersionTLS12,
    }, nil
}

func (c *Config) LoadClientConfig() *tls.Config {
    return &tls.Config{
        InsecureSkipVerify: true, // Note: Only for development
        MinVersion:         tls.VersionTLS12,
    }
}