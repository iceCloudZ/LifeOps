package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleAdminLLMDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	stats, err := s.store.QueryLLMDashboard()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleAdminLLMUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	groupBy := r.URL.Query().Get("groupBy")
	if groupBy == "" {
		groupBy = "day"
	}
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	results, err := s.store.QueryLLMUsage(groupBy, startDate, endDate)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleAdminLLMCostTrend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}
	results, err := s.store.QueryLLMCostTrend(days)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleAdminLLMTopSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	results, err := s.store.QueryTopSessions(limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleAdminLLMRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/dashboard"):
		s.handleAdminLLMDashboard(w, r)
	case strings.HasSuffix(path, "/usage"):
		s.handleAdminLLMUsage(w, r)
	case strings.HasSuffix(path, "/cost-trend"):
		s.handleAdminLLMCostTrend(w, r)
	case strings.HasSuffix(path, "/top-sessions"):
		s.handleAdminLLMTopSessions(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
