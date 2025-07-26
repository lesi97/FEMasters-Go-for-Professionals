package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/lesi97/internal/app"
)

func SetupRoutes(app *app.Application) *chi.Mux {

	routes := chi.NewRouter()
	routes.Get("/health", app.HealthCheck)
	
	routes.Get("/workouts/{id}", app.WorkoutHandler.HandleGetWorkoutById) // chi specific handle for slugs
	routes.Post("/workouts", app.WorkoutHandler.HandleCreateWorkout)
	routes.Put("/workouts/{id}", app.WorkoutHandler.HandleUpdateWorkout)
	routes.Delete("/workouts/{id}", app.WorkoutHandler.HandleDeleteWorkout)

	return routes
}