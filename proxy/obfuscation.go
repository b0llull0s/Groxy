package proxy

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"Groxy/logger"
)

type ObfuscationMode int

const (
	NoObfuscation ObfuscationMode = iota
	HttpHeadersObfuscation
	DomainFrontingSimulation
	CustomJitterObfuscation
)

type TrafficObfuscator struct {
	mode ObfuscationMode
	key  []byte 
}

func NewTrafficObfuscator(mode ObfuscationMode) *TrafficObfuscator {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		logger.Error("Failed to generate encryption key: %v", err)
	}
	
	return &TrafficObfuscator{
		mode: mode,
		key:  key,
	}
}

func (t *TrafficObfuscator) ApplyToRequest(req *http.Request, payload []byte) error {
	switch t.mode {
	case NoObfuscation:
		return nil
		
	case HttpHeadersObfuscation:
		encodedPayload := base64.StdEncoding.EncodeToString(payload)
		chunks := t.chunkString(encodedPayload, 64) 
		
		for i, chunk := range chunks {
			req.Header.Set(fmt.Sprintf("X-Data-%d", i), chunk)
		}
		req.Header.Set("X-Data-Count", fmt.Sprintf("%d", len(chunks)))
		
	case DomainFrontingSimulation:
		if req.Host != "" {
			req.Header.Set("X-Forwarded-For", req.Host)
			req.Host = "cdn.example.com"
			req.Header.Set("Host", "cdn.example.com")
		}
		
		if len(payload) > 0 {
			encodedPayload := base64.StdEncoding.EncodeToString(payload)
			req.AddCookie(&http.Cookie{
				Name:  "session",
				Value: encodedPayload,
			})
		}
		
	case CustomJitterObfuscation:
		encryptedPayload, err := t.encryptData(payload)
		if err != nil {
			return err
		}
		
		size := make([]byte, 4)
		binary.BigEndian.PutUint32(size, uint32(len(encryptedPayload)))
		
		jitterSize := t.randomJitterSize(100, 500)
		jitter := make([]byte, jitterSize)
		rand.Read(jitter)
		
		finalPayload := append(size, jitter...)
		finalPayload = append(finalPayload, encryptedPayload...)
		
		req.Body = io.NopCloser(bytes.NewReader(finalPayload))
		req.ContentLength = int64(len(finalPayload))
		req.Header.Set("Content-Length", fmt.Sprint(len(finalPayload)))
	}
	
	return nil
}

func (t *TrafficObfuscator) ExtractFromResponse(res *http.Response) ([]byte, error) {
	switch t.mode {
	case NoObfuscation:
		return io.ReadAll(res.Body)
		
	case HttpHeadersObfuscation:
		countStr := res.Header.Get("X-Data-Count")
		if countStr == "" {
			return io.ReadAll(res.Body)
		}
		
		var builder strings.Builder
		for i := 0; i < len(res.Header); i++ {
			chunk := res.Header.Get(fmt.Sprintf("X-Data-%d", i))
			if chunk == "" {
				break
			}
			builder.WriteString(chunk)
		}
		
		encodedData := builder.String()
		return base64.StdEncoding.DecodeString(encodedData)
		
	case DomainFrontingSimulation:
		for _, cookie := range res.Cookies() {
			if cookie.Name == "response" {
				return base64.StdEncoding.DecodeString(cookie.Value)
			}
		}
		return io.ReadAll(res.Body)
		
	case CustomJitterObfuscation:
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		
		if len(data) < 4 {
			return data, nil
		}
		
		size := binary.BigEndian.Uint32(data[:4])
		
		totalSize := len(data)
		payloadStart := totalSize - int(size)
		
		if payloadStart < 4 || payloadStart >= totalSize {
			return data, nil
		}
		
		encryptedPayload := data[payloadStart:]
		return t.decryptData(encryptedPayload)
	}
	
	return nil, fmt.Errorf("unknown obfuscation mode")
}

func (t *TrafficObfuscator) chunkString(s string, chunkSize int) []string {
	if len(s) == 0 {
		return []string{}
	}
	chunks := make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		currentLen++
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i+1])
			currentLen = 0
			currentStart = i + 1
		}
	}
	if currentStart < len(s) {
		chunks = append(chunks, s[currentStart:])
	}
	return chunks
}

func (t *TrafficObfuscator) randomJitterSize(min, max int) int {
	maxBigInt := big.NewInt(int64(max - min))
	
	n, err := rand.Int(rand.Reader, maxBigInt)
	if err != nil {
		return min
	}
	
	return min + int(n.Int64())
}

func (t *TrafficObfuscator) encryptData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(t.key)
	if err != nil {
		return nil, err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (t *TrafficObfuscator) decryptData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(t.key)
	if err != nil {
		return nil, err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	
	return gcm.Open(nil, nonce, ciphertext, nil)
}