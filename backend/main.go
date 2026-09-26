package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"minijail/backend/runtime"
)

type Server struct {
	engine *runtime.Engine
	cli    *runtime.CliInterpreter
}

// CORSMiddleware handles headers globally for all requests, including preflight OPTIONS.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding JSON response: %v", err)
	}
}

func errorResponse(w http.ResponseWriter, statusCode int, message string) {
	jsonResponse(w, statusCode, map[string]string{"error": message})
}

func (s *Server) handleContainers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		containers := s.engine.List()
		jsonResponse(w, http.StatusOK, containers)

	case http.MethodPost:
		var req runtime.CreateContainerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
			return
		}
		c, err := s.engine.Create(req)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = s.engine.Start(c.ID)
		jsonResponse(w, http.StatusCreated, c)

	default:
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleContainerDetail(w http.ResponseWriter, r *http.Request) {
	// Cleanly extract path parameters: /api/containers/{id}/{action}
	path := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	
	if id == "" {
		errorResponse(w, http.StatusBadRequest, "Container ID is required")
		return
	}

	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	// Base container operations (No sub-action)
	if action == "" {
		switch r.Method {
		case http.MethodGet:
			c, err := s.engine.Get(id)
			if err != nil {
				errorResponse(w, http.StatusNotFound, err.Error())
				return
			}
			jsonResponse(w, http.StatusOK, c)

		case http.MethodDelete:
			if err := s.engine.Remove(id); err != nil {
				errorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			jsonResponse(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})

		default:
			errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Container state mutations and sub-actions
	switch action {
	case "start", "stop", "pause", "unpause", "kill":
		var err error
		switch action {
		case "start":
			err = s.engine.Start(id)
		case "stop":
			err = s.engine.Stop(id)
		case "pause":
			err = s.engine.Pause(id)
		case "unpause":
			err = s.engine.Unpause(id)
		case "kill":
			err = s.engine.Kill(id)
		}

		if err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		c, _ := s.engine.Get(id)
		jsonResponse(w, http.StatusOK, c)

	case "exec":
		var req runtime.ExecRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "invalid exec payload")
			return
		}
		resp, err := s.engine.Exec(id, req.Command)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, resp)

	case "stress-mem":
		var req struct {
			DeltaMB int64 `json:"deltaMb"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.DeltaMB <= 0 {
			req.DeltaMB = 20
		}
		c, oomKilled, err := s.engine.StressMemory(id, req.DeltaMB)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"container": c,
			"oomKilled": oomKilled,
		})

	case "stress-cpu":
		c, err := s.engine.StressCPU(id, 5)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, c)

	default:
		errorResponse(w, http.StatusNotFound, "unknown container action: "+action)
	}
}

func (s *Server) handleCli(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req runtime.CliCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid CLI command payload")
		return
	}

	resp := s.cli.Execute(req.Command)
	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) handlePrimitives(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, runtime.GetKernelPrimitivesData())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"runtime": "Minijail-Go Engine v1.0",
		"kernel":  "Linux Namespaces + Cgroups v2 Simulation",
	})
}

func main() {
	port := os.Getenv("GO_PORT")
	if port == "" {
		port = "8080"
	}

	engine := runtime.NewEngine()
	cli := runtime.NewCliInterpreter(engine)
	srv := &Server{engine: engine, cli: cli}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", srv.handleHealth)
	mux.HandleFunc("/api/primitives", srv.handlePrimitives)
	mux.HandleFunc("/api/cli", srv.handleCli)
	mux.HandleFunc("/api/containers", srv.handleContainers)
	mux.HandleFunc("/api/containers/", srv.handleContainerDetail)

	// Wrap the default mux with global CORS middleware
	handler := CORSMiddleware(mux)

	addr := "0.0.0.0:" + port
	fmt.Printf("[Minijail-Go] Container Runtime Engine listening on http://%s\n", addr)
	
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}