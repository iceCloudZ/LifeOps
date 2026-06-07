package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

var entityPrefixes = map[string]string{
	"/api/tasks/":          "tasks",
	"/api/events/":         "events",
	"/api/shopping-items/": "shopping-items",
	"/api/notes/":          "notes",
}

func (s *Server) handleEntityRoutes(w http.ResponseWriter, r *http.Request) {
	for prefix, resource := range entityPrefixes {
		if !strings.HasPrefix(r.URL.Path, prefix) {
			continue
		}

		id := strings.TrimPrefix(r.URL.Path, prefix)
		if id == "" {
			switch r.Method {
			case http.MethodGet:
				s.handleListEntities(w, r, resource)
			case http.MethodPost:
				s.handleCreateEntity(w, r, resource)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			s.handleGetEntity(w, r, resource, id)
		case http.MethodPut:
			s.handleUpdateEntity(w, r, resource, id)
		case http.MethodDelete:
			s.handleDeleteEntity(w, r, resource, id)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	writeJSONError(w, http.StatusNotFound, "not_found", "resource not found")
}

func (s *Server) handleListEntities(w http.ResponseWriter, r *http.Request, resource string) {
	var results interface{}
	var err error

	switch resource {
	case "tasks":
		results, err = s.store.ListTasks()
	case "events":
		results, err = s.store.ListEvents()
	case "shopping-items":
		results, err = s.store.ListShoppingItems()
	case "notes":
		results, err = s.store.ListNotes()
	default:
		writeJSONError(w, http.StatusNotFound, "not_found", "resource not found")
		return
	}

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to list")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleCreateEntity(w http.ResponseWriter, r *http.Request, resource string) {
	now := time.Now().UTC().Format(time.RFC3339)
	id := newID()

	switch resource {
	case "tasks":
		var req struct {
			Title       string  `json:"title"`
			Description string  `json:"description"`
			DueAt       *string `json:"due_at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "title is required")
			return
		}
		task := &FamilyTask{
			ID: id, Title: req.Title, Description: req.Description,
			DueAt: req.DueAt, Status: "open", CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.CreateTask(task); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "create failed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)

	case "events":
		var req struct {
			Title       string  `json:"title"`
			Description string  `json:"description"`
			StartsAt    *string `json:"starts_at"`
			EndsAt      *string `json:"ends_at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "title is required")
			return
		}
		event := &FamilyEvent{
			ID: id, Title: req.Title, Description: req.Description,
			StartsAt: req.StartsAt, EndsAt: req.EndsAt,
			CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.CreateEvent(event); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "create failed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(event)

	case "shopping-items":
		var req struct {
			Name     string `json:"name"`
			Quantity string `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "name is required")
			return
		}
		item := &ShoppingItem{
			ID: id, Name: req.Name, Quantity: req.Quantity,
			Status: "open", CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.CreateShoppingItem(item); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "create failed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(item)

	case "notes":
		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Topic   string `json:"topic"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "title is required")
			return
		}
		note := &FamilyNote{
			ID: id, Title: req.Title, Content: req.Content,
			Topic: req.Topic, CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.CreateNote(note); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "create failed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(note)

	default:
		writeJSONError(w, http.StatusNotFound, "not_found", "resource not found")
	}
}

func (s *Server) handleGetEntity(w http.ResponseWriter, r *http.Request, resource, id string) {
	switch resource {
	case "tasks":
		t, err := s.store.GetTask(id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "query failed")
			return
		}
		if t == nil {
			writeJSONError(w, http.StatusNotFound, "not_found", "task not found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)

	case "events":
		e, err := s.store.GetEvent(id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "query failed")
			return
		}
		if e == nil {
			writeJSONError(w, http.StatusNotFound, "not_found", "event not found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(e)

	case "shopping-items":
		si, err := s.store.GetShoppingItem(id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "query failed")
			return
		}
		if si == nil {
			writeJSONError(w, http.StatusNotFound, "not_found", "shopping item not found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(si)

	case "notes":
		n, err := s.store.GetNote(id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "query failed")
			return
		}
		if n == nil {
			writeJSONError(w, http.StatusNotFound, "not_found", "note not found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(n)

	default:
		writeJSONError(w, http.StatusNotFound, "not_found", "resource not found")
	}
}

func (s *Server) handleUpdateEntity(w http.ResponseWriter, r *http.Request, resource, id string) {
	now := time.Now().UTC().Format(time.RFC3339)

	switch resource {
	case "tasks":
		var req struct {
			Title       string  `json:"title"`
			Description string  `json:"description"`
			DueAt       *string `json:"due_at"`
			Status      string  `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if err := s.store.UpdateTask(id, req.Title, req.Description, req.DueAt, req.Status, now); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "update failed")
			return
		}
		result, _ := s.store.GetTask(id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	case "events":
		var req struct {
			Title       string  `json:"title"`
			Description string  `json:"description"`
			StartsAt    *string `json:"starts_at"`
			EndsAt      *string `json:"ends_at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if err := s.store.UpdateEvent(id, req.Title, req.Description, req.StartsAt, req.EndsAt, now); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "update failed")
			return
		}
		result, _ := s.store.GetEvent(id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	case "shopping-items":
		var req struct {
			Name     string `json:"name"`
			Quantity string `json:"quantity"`
			Status   string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if err := s.store.UpdateShoppingItem(id, req.Name, req.Quantity, req.Status, now); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "update failed")
			return
		}
		result, _ := s.store.GetShoppingItem(id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	case "notes":
		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Topic   string `json:"topic"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if err := s.store.UpdateNote(id, req.Title, req.Content, req.Topic, now); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "update failed")
			return
		}
		result, _ := s.store.GetNote(id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	default:
		writeJSONError(w, http.StatusNotFound, "not_found", "resource not found")
	}
}

func (s *Server) handleDeleteEntity(w http.ResponseWriter, r *http.Request, resource, id string) {
	var err error

	switch resource {
	case "tasks":
		err = s.store.DeleteTask(id)
	case "events":
		err = s.store.DeleteEvent(id)
	case "shopping-items":
		err = s.store.DeleteShoppingItem(id)
	case "notes":
		err = s.store.DeleteNote(id)
	default:
		writeJSONError(w, http.StatusNotFound, "not_found", "resource not found")
		return
	}

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "delete failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
