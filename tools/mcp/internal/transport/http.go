package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

// HTTPTransport handles HTTP-based MCP communication
type HTTPTransport struct {
	port       int
	mcpHandler MCPHandler
	logger     *logrus.Logger
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(port int, mcpHandler MCPHandler, logger *logrus.Logger) *HTTPTransport {
	return &HTTPTransport{
		port:       port,
		mcpHandler: mcpHandler,
		logger:     logger,
	}
}

// Start starts the HTTP server
func (h *HTTPTransport) Start() error {
	mux := http.NewServeMux()
	
	// MCP protocol endpoint
	mux.HandleFunc("/mcp", h.handleMCPRequest)
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	
	addr := fmt.Sprintf(":%d", h.port)
	h.logger.Infof("Starting HTTP MCP server on %s", addr)
	
	return http.ListenAndServe(addr, mux)
}

// handleMCPRequest handles MCP JSON-RPC requests over HTTP
func (h *HTTPTransport) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, MCP-Protocol-Version")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Check MCP protocol version
	protocolVersion := r.Header.Get("MCP-Protocol-Version")
	if protocolVersion == "" {
		protocolVersion = "2024-11-05" // Default version
	}
	
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	h.logger.Debugf("Received MCP request: %s", string(body))
	
	// Handle the request using MCPHandler interface
	response, err := h.mcpHandler.HandleMessage(r.Context(), body)
	if err != nil {
		h.logger.Errorf("Failed to handle MCP message: %v", err)
		http.Error(w, "Failed to handle MCP message", http.StatusInternalServerError)
		return
	}
	
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("MCP-Protocol-Version", protocolVersion)
	
	// Send response
	if _, err := w.Write(response); err != nil {
		h.logger.Errorf("Failed to write response: %v", err)
		return
	}
	
	h.logger.Debugf("Sent MCP response: %s", string(response))
}

// MCPHandler interface for handling MCP requests
type MCPHandler interface {
	HandleMessage(ctx context.Context, message []byte) ([]byte, error)
}