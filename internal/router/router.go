package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/lesi97/internal/app"
)

func SetupRoutes(app *app.Application) *chi.Mux {

	routes := chi.NewRouter()
	routes.Group(func (r chi.Router) {
		r.Use(app.Middleware.Authenticate)

		r.Get("/workouts/{id}", app.Middleware.RequireUser(app.WorkoutHandler.HandleGetWorkoutById)) // {id} is chi specific handle for slugs
		r.Post("/workouts", app.Middleware.RequireUser(app.WorkoutHandler.HandleCreateWorkout))
		r.Put("/workouts/{id}", app.Middleware.RequireUser(app.WorkoutHandler.HandleUpdateWorkout))
		r.Delete("/workouts/{id}", app.Middleware.RequireUser(app.WorkoutHandler.HandleDeleteWorkout))
	})


	routes.Get("/health", app.HealthCheck)
	
	routes.Post("/users", app.UserHandler.HandleRegisterUser)
	routes.Post("/tokens/authentication", app.TokenHandler.HandleCreateToken)

	return routes
}