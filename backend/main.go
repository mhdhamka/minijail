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

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	enableCORS(w)
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
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
		// Optionally auto-start if created
		_ = s.engine.Start(c.ID)
		jsonResponse(w, http.StatusCreated, c)

	default:
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleContainerDetail(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Path format: /api/containers/{id}/{action}
	path := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	parts := strings.Split(path, "/")
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

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

	// Actions: start, stop, pause, unpause, kill, exec, stress-mem, stress-cpu
	switch action {
	case "start":
		if err := s.engine.Start(id); err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		c, _ := s.engine.Get(id)
		jsonResponse(w, http.StatusOK, c)

	case "stop":
		if err := s.engine.Stop(id); err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		c, _ := s.engine.Get(id)
		jsonResponse(w, http.StatusOK, c)

	case "pause":
		if err := s.engine.Pause(id); err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		c, _ := s.engine.Get(id)
		jsonResponse(w, http.StatusOK, c)

	case "unpause":
		if err := s.engine.Unpause(id); err != nil {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		c, _ := s.engine.Get(id)
		jsonResponse(w, http.StatusOK, c)

	case "kill":
		if err := s.engine.Kill(id); err != nil {
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
		type StressMemReq struct {
			DeltaMB int64 `json:"deltaMb"`
		}
		var req StressMemReq
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
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	primitives := runtime.GetKernelPrimitivesData()
	jsonResponse(w, http.StatusOK, primitives)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
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

	addr := "0.0.0.0:" + port
	fmt.Printf("[Minijail-Go] Container Runtime Engine listening on http://%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
