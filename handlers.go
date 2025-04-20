package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

// global JWT setup
var tokenAuth *jwtauth.JWTAuth

func init() {
	// replace with a secure key from env in prod
	tokenAuth = jwtauth.New("HS256", []byte("secret-signing-key"), nil)
}

type errorResp struct {
	Error string `json:"error"`
}

var keyMatchRegex = regexp.MustCompile(`\"(\w+)\":`)
var wordBarrierRegex = regexp.MustCompile(`(\w{2,})([A-Z])`)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	marshalled, _ := json.Marshal(v)

	converted := keyMatchRegex.ReplaceAllFunc(
		marshalled,
		func(match []byte) []byte {
			return bytes.ToLower(wordBarrierRegex.ReplaceAll(
				match,
				[]byte(`${1}_${2}`),
			))
		},
	)

	_, _ = w.Write(converted)
}

// --- Auth Handlers ---

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err := CreateUser(db, req.Username, string(hash)); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
	}
}

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}
		userID, pwHash, err := GetUserByUsername(db, req.Username)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResp{"invalid credentials"})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(pwHash), []byte(req.Password)) != nil {
			writeJSON(w, http.StatusUnauthorized, errorResp{"invalid credentials"})
			return
		}
		// create token
		_, tokenString, _ := tokenAuth.Encode(map[string]any{
			"user_id": userID,
			"exp":     jwtauth.ExpireIn(time.Hour * 24 * 30),
		})
		writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
	}
}

func ProjectAccess(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Token   string `json:"token"`
			Action  string `json:"action"`
			TableId int    `json:"table"`
			VarName string `json:"variable"`
			Value   string `json:"value,omitempty"`
			Type    string `json:"type,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}
		_, _, err := GetProjectByToken(db, req.Token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResp{"invalid project token"})
			return
		}

		switch req.Action {
		case "get":
			val, typ, err := GetVariable(db, req.TableId, req.VarName)
			if err == ErrNotFound {
				writeJSON(w, http.StatusNotFound, errorResp{"variable not found"})
				return
			} else if err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"value": val,
				"type":  typ,
			})

		case "set":
			if req.Type == "" {
				writeJSON(w, http.StatusBadRequest, errorResp{"type is required"})
				return
			}
			if err := SetVariable(db, req.TableId, req.VarName, req.Value, req.Type); err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

		default:
			writeJSON(w, http.StatusBadRequest, errorResp{"unknown action"})
		}
	}
}

func ProjectCreate(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}

		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))

		token := uuid.New().String()

		pid, err := CreateProject(db, userID, req.Name, token)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]int{"project_id": pid})
	}
}

func ProjectList(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))
		projects, err := ListProjects(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, projects)
	}
}

func ProjectRename(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}

		projectIdStr := chi.URLParam(r, "projectID")
		projectId, err := strconv.Atoi(projectIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid project ID"})
		}

		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))

		if err := RenameProject(db, projectId, req.Name, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func ProjectDelete(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the JWT token
		projectIdStr := chi.URLParam(r, "projectID")
		projectId, err := strconv.Atoi(projectIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid project ID"})
		}
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))

		if err := DeleteProject(db, projectId, userID); err != nil {

			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func TableCreate(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}

		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))
		projectIdStr := chi.URLParam(r, "projectID")
		projectId, err := strconv.Atoi(projectIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid project ID"})
		}

		tid, err := CreateTable(db, projectId, req.Name, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]int{"table_id": tid})
	}
}

func TableList(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))
		projectIdStr := chi.URLParam(r, "projectID")
		projectId, err := strconv.Atoi(projectIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid project ID"})
		}

		tables, err := ListTables(db, projectId, userID)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, tables)
	}
}

func TableRename(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}

		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))
		tableIdStr := chi.URLParam(r, "tableID")
		tableId, err := strconv.Atoi(tableIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid table ID"})
		}

		if err := RenameTable(db, tableId, req.Name, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func TableDelete(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))
		tableIdStr := chi.URLParam(r, "tableID")
		tableId, err := strconv.Atoi(tableIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid table ID"})
		}

		if err := DeleteTable(db, tableId, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func VariableList(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))
		tableIdStr := chi.URLParam(r, "tableID")
		tableId, err := strconv.Atoi(tableIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid table ID"})
		}

		variables, err := ListVariables(db, tableId, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, variables)
	}
}

func VariableCreate(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name  string `json:"name"`
			Value string `json:"value"`
			Type  string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}

		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))
		tableIdStr := chi.URLParam(r, "tableID")
		tableId, err := strconv.Atoi(tableIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid table ID"})
		}

		if err := CreateVariable(db, tableId, req.Name, req.Value, req.Type, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func VariableDelete(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))
		tableIdStr := chi.URLParam(r, "tableID")
		tableId, err := strconv.Atoi(tableIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid table ID"})
		}
		name := chi.URLParam(r, "name")

		if err := DeleteVariable(db, tableId, name, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func VariableUpdate(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			NewName string `json:"new_name"`
			Type    string `json:"new_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{err.Error()})
			return
		}

		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))
		tableIdStr := chi.URLParam(r, "tableID")
		tableId, err := strconv.Atoi(tableIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid table ID"})
		}
		name := chi.URLParam(r, "name")

		if err := UpdateVariable(db, tableId, name, req.NewName, req.Type, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func ProjectLoad(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectIdStr := chi.URLParam(r, "projectID")
		projectId, err := strconv.Atoi(projectIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResp{"invalid project ID"})
		}

		// Get the user ID from the JWT token
		_, claims, _ := jwtauth.FromContext(r.Context())
		userID := int(claims["user_id"].(float64))

		project, err := GetProject(db, projectId, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, project)
	}
}

func MountAPIRoutes(r chi.Router, db *sql.DB) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
	}))

	// Public API routes
	r.Route("/api", func(r chi.Router) {
		r.Post("/register", Register(db))
		r.Post("/login", Login(db))
		r.Post("/access", ProjectAccess(db))

		// JWT‑protected subrouter:
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))

			r.Use(render.SetContentType(render.ContentTypeJSON))
			r.Route("/projects", func(r chi.Router) {
				r.Post("/", ProjectCreate(db))
				r.Get("/", ProjectList(db))
				r.Route("/{projectID}", func(r chi.Router) {
					r.Get("/", ProjectLoad(db))
					r.Put("/", ProjectRename(db))
					r.Delete("/", ProjectDelete(db))

					r.Route("/tables", func(r chi.Router) {
						r.Post("/", TableCreate(db))
						r.Get("/", TableList(db))
						r.Route("/{tableID}", func(r chi.Router) {
							r.Put("/", TableRename(db))
							r.Delete("/", TableDelete(db))

							r.Route("/variables", func(r chi.Router) {
								r.Post("/", VariableCreate(db))
								r.Get("/", VariableList(db))
								r.Route("/{name}", func(r chi.Router) {
									r.Put("/", VariableUpdate(db))
									r.Delete("/", VariableDelete(db))
								})
							})
						})
					})
				})
			})
		})
	})

	// Frontend routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/index.html")
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/login.html")
	})

	r.Get("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/register.html")
	})

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "frontend/dashboard.html")
		})
		r.Route("/project/{projectID}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "frontend/project.html")
			})
		})
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))
}
