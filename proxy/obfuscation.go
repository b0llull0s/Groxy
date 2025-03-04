package proxy

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
//	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"net/http"
//	"strings"
	"time"

	"Groxy/logger"
)

type ObfuscationMode int

const (
	StrongObfuscation ObfuscationMode = iota
)

type TrafficObfuscator struct {
	requestKey  []byte
	responseKey []byte
	hmacKey     []byte
}

func NewTrafficObfuscator() *TrafficObfuscator {
	requestKey := make([]byte, 32)
	responseKey := make([]byte, 32)
	hmacKey := make([]byte, 64)
	
	if _, err := rand.Read(requestKey); err != nil {
		logger.Error("Failed to generate request encryption key: %v", err)
	}
	if _, err := rand.Read(responseKey); err != nil {
		logger.Error("Failed to generate response encryption key: %v", err)
	}
	if _, err := rand.Read(hmacKey); err != nil {
		logger.Error("Failed to generate HMAC key: %v", err)
	}
	
	return &TrafficObfuscator{
		requestKey:  requestKey,
		responseKey: responseKey,
		hmacKey:     hmacKey,
	}
}

func (t *TrafficObfuscator) ApplyToRequest(req *http.Request) error {
	if req.Body == nil {
		req.Body = io.NopCloser(bytes.NewReader([]byte{}))
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body.Close()

	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(time.Now().UnixNano()))
	combinedData := append(timestamp, bodyBytes...)

	encryptedBody, err := t.encryptData(combinedData, t.requestKey)
	if err != nil {
		return err
	}

	hmac := t.generateHMAC(encryptedBody)
	finalPayload := append(hmac, encryptedBody...)
	jitteredPayload := t.addJitter(finalPayload)

	newReq, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewReader(jitteredPayload))
	if err != nil {
		return err
	}

	newReq.Host = req.Host
	t.obfuscateHeaders(newReq)
	*req = *newReq

	return nil
}

func (t *TrafficObfuscator) obfuscateHeaders(req *http.Request) {
	newHeaders := make(http.Header)

	noiseHeaders := []string{
		"X-Proxy-Token",
		"X-Connection-Hash",
		"X-Routing-Key",
		"X-Timestamp",
	}

	for _, header := range noiseHeaders {
		newHeaders.Set(header, t.generateRandomString(32))
	}

	newHeaders.Set("Content-Type", "application/octet-stream")

	req.Header = newHeaders
}

func (t *TrafficObfuscator) ExtractFromResponse(res *http.Response) ([]byte, error) {
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body.Close()

	dejitteredBody := t.removeJitter(bodyBytes)

	if len(dejitteredBody) < sha256.Size+1 {
		return nil, fmt.Errorf("response too short for decryption")
	}

	receivedHmac := dejitteredBody[:sha256.Size]
	encryptedData := dejitteredBody[sha256.Size:]

	if !t.verifyHMAC(encryptedData, receivedHmac) {
		return nil, fmt.Errorf("HMAC verification failed")
	}

	decryptedBody, err := t.decryptData(encryptedData, t.responseKey)
	if err != nil {
		return nil, err
	}

	if len(decryptedBody) < 8 {
		return nil, fmt.Errorf("decrypted data too short")
	}
	return decryptedBody[8:], nil
}

func (t *TrafficObfuscator) generateHMAC(data []byte) []byte {
	mac := hmac.New(sha256.New, t.hmacKey)
	mac.Write(data)
	return mac.Sum(nil)
}

func (t *TrafficObfuscator) verifyHMAC(data []byte, expectedHmac []byte) bool {
	mac := hmac.New(sha256.New, t.hmacKey)
	mac.Write(data)
	return hmac.Equal(mac.Sum(nil), expectedHmac)
}

func (t *TrafficObfuscator) encryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
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

func (t *TrafficObfuscator) decryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
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

func (t *TrafficObfuscator) addJitter(data []byte) []byte {
	jitterSize := t.randomJitterSize(100, 500)
	jitter := make([]byte, jitterSize)
	rand.Read(jitter)
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(data)))

	return append(append(size, jitter...), data...)
}

func (t *TrafficObfuscator) removeJitter(data []byte) []byte {
	if len(data) < 4 {
		return data
	}
	
	size := binary.BigEndian.Uint32(data[:4])
	
	if len(data) < int(4 + size) {
		return data
	}
	return data[4+len(data)-int(size):]
}

func (t *TrafficObfuscator) randomJitterSize(min, max int) int {
	maxBigInt := big.NewInt(int64(max - min))
	
	n, err := rand.Int(rand.Reader, maxBigInt)
	if err != nil {
		return min
	}
	return min + int(n.Int64())
}

func (t *TrafficObfuscator) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}