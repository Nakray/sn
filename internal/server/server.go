package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Nakray/sn/internal/config"
	"github.com/Nakray/sn/internal/database"
	"github.com/Nakray/sn/internal/monitoring"
	"github.com/gorilla/mux"
)

type Server struct {
	db         *database.DB
	monitoring *monitoring.Service
	config     *config.Config
	router     *mux.Router
}

func New(db *database.DB, mon *monitoring.Service, cfg *config.Config) *Server {
	s := &Server{
		db:         db,
		monitoring: mon,
		config:     cfg,
		router:     mux.NewRouter(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Static UI
	s.router.HandleFunc("/", s.handleIndex).Methods("GET")

	// Monitoring tasks API
	s.router.HandleFunc("/api/tasks", s.handleGetTasks).Methods("GET")
	s.router.HandleFunc("/api/tasks", s.handleCreateTask).Methods("POST")
	s.router.HandleFunc("/api/tasks/{id}", s.handleDeleteTask).Methods("DELETE")

	// Accounts API
	s.router.HandleFunc("/api/accounts", s.handleGetAccounts).Methods("GET")
	s.router.HandleFunc("/api/accounts", s.handleCreateAccount).Methods("POST")
	s.router.HandleFunc("/api/accounts/{id}", s.handleDeleteAccount).Methods("DELETE")
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Server.Port)
	log.Printf("HTTP server listening on %s\n", addr)
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, indexHTML)
}

func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.ListMonitoringTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task database.MonitoringTask
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.db.CreateMonitoringTask(&task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteMonitoringTask(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.db.ListAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var account database.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.db.CreateAccount(&account); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteAccount(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
