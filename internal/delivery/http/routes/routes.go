package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"task-management-system/internal/delivery/http/handlers"
	"task-management-system/internal/delivery/http/middleware"
	"task-management-system/internal/usecase"
)

// NewRouter creates a new HTTP router
func NewRouter(
	taskUseCase *usecase.TaskUseCase,
	userUseCase *usecase.UserUseCase,
	authUseCase *usecase.AuthUseCase,
) http.Handler {
	// Create router
	router := mux.NewRouter()

	// Create handlers
	taskHandler := handlers.NewTaskHandler(taskUseCase)
	userHandler := handlers.NewUserHandler(userUseCase)
	authHandler := handlers.NewAuthHandler(authUseCase, userUseCase)

	// Apply global middlewares
	router.Use(middleware.Recover)
	router.Use(middleware.Logger)
	router.Use(middleware.CORS)

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Auth routes (no authentication required)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", authHandler.Register).Methods("POST")
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")
	auth.HandleFunc("/refresh-token", authHandler.RefreshToken).Methods("POST")

	// Routes that require authentication
	authenticated := api.NewRoute().Subrouter()
	authenticated.Use(middleware.Auth(authUseCase))

	// User routes
	authenticated.HandleFunc("/me", userHandler.GetProfile).Methods("GET")
	authenticated.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	authenticated.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")

	// Task routes
	authenticated.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
	authenticated.HandleFunc("/tasks", taskHandler.ListTasks).Methods("GET")
	authenticated.HandleFunc("/tasks/{id}", taskHandler.GetTask).Methods("GET")
	authenticated.HandleFunc("/tasks/{id}", taskHandler.UpdateTask).Methods("PUT")
	authenticated.HandleFunc("/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")
	authenticated.HandleFunc("/tasks/{id}/assign", taskHandler.AssignTask).Methods("POST")
	authenticated.HandleFunc("/users/{id}/tasks", taskHandler.GetUserTasks).Methods("GET")

	// Health check route (no authentication required)
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	return router
}
