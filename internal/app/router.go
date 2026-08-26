package app

import (
	"net/http"
	"todo/internal/auth/handlers"
	"todo/internal/jsonrender"
	"todo/internal/middlewares"
	todohandlers "todo/internal/todo/handlers"
	usershandlers "todo/internal/users/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(taskHandler *todohandlers.TaskHandler, usersHandler *handlers.AuthHandler, authMiddleware *middlewares.Auth, corsMiddleware *middlewares.CORS, profileHandler *usershandlers.ProfileHandler) *chi.Mux {
	mainRouter := chi.NewRouter()

	mainRouter.Use(middleware.StripSlashes)
	mainRouter.Use(corsMiddleware.CORSMiddleware)

	mainRouter.Route("/tasks", func(r chi.Router) {
		r.Use(authMiddleware.AuthMiddleware)
		r.Get("/", taskHandler.GetAllTasksHandler)
		r.Post("/", taskHandler.CreateTaskHandler)

		r.Route("/{task_id}", func(r chi.Router) {
			r.Get("/", taskHandler.GetTaskHandler)
			r.Delete("/", taskHandler.DeleteTaskHandler)
			r.Patch("/", taskHandler.UpdateTaskPatch)
			r.Put("/", taskHandler.UpdateTaskPut)
		})

	})

	mainRouter.Post("/register", usersHandler.Register)
	mainRouter.Post("/login", usersHandler.Login)
	mainRouter.Post("/refresh", usersHandler.Refresh)
	mainRouter.Post("/logout", usersHandler.Logout)

	mainRouter.Route("/me", func(r chi.Router) {
		r.Use(authMiddleware.AuthMiddleware)
		r.Get("/", profileHandler.GetProfile)
	})

	mainRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		jsonrender.JSONResponse(map[string]any{"detail": "Привет! Это TODO API"}, w, 200)
	})

	mainRouter.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		jsonrender.JSONResponse(map[string]any{"detail": "Method Not Allowed"}, w, 405)
	})

	mainRouter.NotFound(func(w http.ResponseWriter, r *http.Request) {
		jsonrender.JSONResponse(map[string]any{"detail": "Page Not Found"}, w, 404)
	})

	return mainRouter
}
