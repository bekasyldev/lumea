package main

import (
	"log"
	"net/http"
	"time"

	"github.com/bekasyldev/lumea/internal/auth"
	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/bekasyldev/lumea/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
)

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID) // rate limiting
	r.Use(middleware.RealIP)    // rate limit || analytics || tracing
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // recover from crashes

	queries := database.New(app.db) // store connection inside the Queries struct
	userRepo := user.NewRepository(queries)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	authRepo := auth.NewRepository(queries)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	// public routes
	r.Group(func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Server is running :)"))
		})
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// private routes (require a valid JWT)
	r.Group(func(r chi.Router) {
		r.Use(app.requireAuth)
		r.Get("/users", userHandler.List)
		r.Get("/users/{id}", userHandler.GetByID)
	})

	return r
}

func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("server has started at add %s", app.config.addr)

	return srv.ListenAndServe()
}

type application struct {
	config config
	db     *pgx.Conn
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}
