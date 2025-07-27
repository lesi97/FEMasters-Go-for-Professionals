package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lesi97/internal/api"
	"github.com/lesi97/internal/middleware"
	"github.com/lesi97/internal/store"
	"github.com/lesi97/migrations"
)

type Application struct {
	DB 				*sql.DB
	Logger 			*log.Logger
	Middleware 		middleware.UserMiddleware
	WorkoutHandler 	*api.WorkoutHandler
	UserHandler 	*api.UserHandler
	TokenHandler 	*api.TokenHandler
}

func NewApplication() (*Application, error) {

	pgDB, err := store.Open()
	if err != nil {
		return nil, err
	}

	err = store.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	workoutStore := store.NewPostgresWorkoutStore(pgDB)
	userStore := store.NewPostgresUserStore(pgDB)
	tokenStore := store.NewPostgresTokenStore(pgDB)

	middlewareHandler := middleware.UserMiddleware{UserStore: userStore, Logger: logger}
	workoutHandler := api.NewWorkoutHandler(workoutStore, logger)
	userHandler := api.NewUserHandler(userStore, logger)
	tokenHandler := api.NewTokenHandler(tokenStore, userStore, logger)

	app := &Application{
		DB: pgDB,
		Logger: logger,
		Middleware: middlewareHandler,
		WorkoutHandler: workoutHandler,
		UserHandler: userHandler,
		TokenHandler: tokenHandler,
	}

	return app, nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Status is available\n") // Fprint allows response to writer??!??!?
	fmt.Println("Status is available")
}