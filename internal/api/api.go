package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// API represents the QuickBot REST API
type API struct {
	agent    *Agent
	memory   *Memory
	scheduler *Scheduler
	port     int
}

// NewAPI creates a new API instance
func NewAPI(agent *Agent, memory *Memory, scheduler *Scheduler, port int) *API {
	return &API{
		agent:     agent,
		memory:    memory,
		scheduler: scheduler,
		port:      port,
	}
}

// Start starts the API server
func (a *API) Start() error {
	// Register routes
	http.HandleFunc("/", a.handleRoot)
	http.HandleFunc("/health", a.handleHealth)
	http.HandleFunc("/api/v1/chat", a.handleChat)
	http.HandleFunc("/api/v1/memory/", a.handleMemory)
	http.HandleFunc("/api/v1/tasks", a.handleTasks)
	http.HandleFunc("/api/v1/status", a.handleStatus)

	// Start server
	addr := fmt.Sprintf(":%d", a.port)
	log.Printf("API server starting on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  - GET  /health")
	log.Printf("  - POST /api/v1/chat")
	log.Printf("  - GET  /api/v1/memory/<key>")
	log.Printf("  - POST /api/v1/memory")
	log.Printf("  - GET  /api/v1/tasks")
	log.Printf("  - GET  /api/v1/status")

	return http.ListenAndServe(addr, nil)
}

// Response represents API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// handleRoot handles root endpoint
func (a *API) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Success: true,
		Data: map[string]string{
			"name":    "QuickBot API",
			"version": "1.0.0",
			"docs":    "See README.md for API documentation",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleHealth handles health check endpoint
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.sendMethodNotAllowed(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"services": map[string]bool{
				"api":       true,
				"memory":    a.memory != nil,
				"scheduler": a.scheduler != nil,
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleChat handles chat endpoint
func (a *API) handleChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		a.sendMethodNotAllowed(w)
		return
	}

	var request struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		a.sendError(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if request.Message == "" {
		a.sendError(w, "Message is required")
		return
	}

	if request.SessionID == "" {
		request.SessionID = fmt.Sprintf("api_%d", time.Now().UnixNano())
	}

	// Process message
	responseData, err := a.agent.ProcessMessage(request.SessionID, request.Message)
	if err != nil {
		a.sendError(w, fmt.Sprintf("Failed to process: %v", err))
		return
	}

	response := Response{
		Success: true,
		Data: map[string]interface{}{
			"response":   responseData,
			"session_id": request.SessionID,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleMemory handles memory operations
func (a *API) handleMemory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Extract key from URL
		key := r.URL.Path[len("/api/v1/memory/"):]
		if key == "" {
			a.sendError(w, "Key is required")
			return
		}

		// Get memory value
		value, err := a.agent.GetMemory(key)
		if err != nil {
			a.sendError(w, fmt.Sprintf("Failed to get memory: %v", err))
			return
		}

		response := Response{
			Success: true,
			Data: map[string]interface{}{
				"key":   key,
				"value": value,
			},
		}

		json.NewEncoder(w).Encode(response)

	case http.MethodPost:
		var request struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			a.sendError(w, fmt.Sprintf("Invalid request: %v", err))
			return
		}

		if request.Key == "" || request.Value == "" {
			a.sendError(w, "Key and value are required")
			return
		}

		// Set memory value
		err = a.agent.SetMemory(request.Key, request.Value)
		if err != nil {
			a.sendError(w, fmt.Sprintf("Failed to set memory: %v", err))
			return
		}

		response := Response{
			Success: true,
			Data: map[string]interface{}{
				"action": "set",
				"key":    request.Key,
			},
		}

		json.NewEncoder(w).Encode(response)

	default:
		a.sendMethodNotAllowed(w)
	}
}

// handleTasks handles tasks endpoint
func (a *API) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		a.sendMethodNotAllowed(w)
		return
	}

	tasks := a.scheduler.GetAllTasks()

	response := Response{
		Success: true,
		Data: map[string]interface{}{
			"count": len(tasks),
			"tasks": tasks,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleStatus handles status endpoint
func (a *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.sendMethodNotAllowed(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Success: true,
		Data: map[string]interface{}{
			"bot":        a.agent.Config().Bot.Name,
			"ai_provider": a.agent.Config().AI.Provider,
			"ai_model":   a.agent.Config().AI.Model,
			"uptime":     time.Since(time.Now()).String(),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// sendError sends error response
func (a *API) sendError(w http.ResponseWriter, message string) {
	response := Response{
		Success: false,
		Error:   message,
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

// sendMethodNotAllowed sends method not allowed response
func (a *API) sendMethodNotAllowed(w http.ResponseWriter) {
	response := Response{
		Success: false,
		Error:   "Method not allowed",
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(response)
}

// TestAPI tests the API module
func TestAPI() {
	log.Println("Testing API module...")

	// Create test components
	config := GetDefaultConfig()
	memory, _ := NewMemory("test_api_memory.db", 100)
	scheduler, _ := NewScheduler("test_api_scheduler.db")
	agent := NewSimpleAgent(config, NewSimpleMemory(100), NewSimpleScheduler())

	// Create API instance
	api := NewAPI(nil, nil, nil, 8080)

	log.Println("✓ API module initialized")

	// Cleanup
	memory.Close()
	scheduler.Close()

	os.Remove("test_api_memory.db")
	os.Remove("test_api_scheduler.db")

	log.Println("✓ API module tests passed")
}
