package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type CreateMemberRequest struct {
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	BirthDate *string `json:"birth_date"`
}

type UpdateMemberRequest struct {
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	BirthDate *string `json:"birth_date"`
}

func (s *Server) handleMembers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		members, err := s.store.ListFamilyMembers()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to list members")
			return
		}
		if members == nil {
			members = []FamilyMember{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(members)
	case http.MethodPost:
		var req CreateMemberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "name is required")
			return
		}
		member := &FamilyMember{
			ID:        newID(),
			Name:      strings.TrimSpace(req.Name),
			Role:      strings.TrimSpace(req.Role),
			BirthDate: req.BirthDate,
		}
		if err := s.store.CreateFamilyMember(member); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to create member")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(member)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleMemberByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/members/")
	if id == "" {
		writeJSONError(w, http.StatusNotFound, "not_found", "member not found")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req UpdateMemberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "name is required")
			return
		}
		if err := s.store.UpdateFamilyMember(id, strings.TrimSpace(req.Name), strings.TrimSpace(req.Role), req.BirthDate); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to update member")
			return
		}
		result, err := s.store.GetFamilyMember(id)
		if err != nil || result == nil {
			writeJSONError(w, http.StatusNotFound, "not_found", "member not found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	case http.MethodDelete:
		if err := s.store.DeleteFamilyMember(id); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to delete member")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
