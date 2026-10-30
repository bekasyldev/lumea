package main

import (
	"log"
	"net/http"
	"time"

	"github.com/bekasyldev/lumea/internal/auth"
	"github.com/bekasyldev/lumea/internal/categories"
	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/bekasyldev/lumea/internal/media"
	"github.com/bekasyldev/lumea/internal/media/s3store"
	"github.com/bekasyldev/lumea/internal/orders"
	"github.com/bekasyldev/lumea/internal/product"
	"github.com/bekasyldev/lumea/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
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

	productRepo := product.NewRepository(queries)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)

	categoryRepo := categories.NewRepository(queries)
	categoryService := categories.NewService(categoryRepo)
	categoryHandler := categories.NewHandler(categoryService)

	orderRepo := orders.NewRepository(app.db, queries)
	orderService := orders.NewService(orderRepo)
	orderHandler := orders.NewHandler(orderService)

	mediaRepo := media.NewRepository(queries)
	mediaService := media.NewService(app.storage, mediaRepo)
	mediaHandler := media.NewHandler(mediaService)

	// public routes
	r.Group(func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Server is running :)"))
		})
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.RefreshToken)

		r.Get("/products", productHandler.ListProducts)
		r.Get("/products/{id}/variants", productHandler.ListProductVariants)
		r.Get("/products/{id}/images", mediaHandler.GetListProductImages)
		r.Get("/products/{id}/variants/{variantId}/images", mediaHandler.GetListVariantImages)
		r.Get("/categories", categoryHandler.ListCategories)
		r.Get("/categories/{id}", categoryHandler.GetCategoryByID)
		r.Get("/categories/slug/{slug}", categoryHandler.GetCategoryBySlug)

	})

	// private routes (require a valid JWT)
	r.Group(func(r chi.Router) {
		r.Use(app.requireAuth)
		r.Get("/users", userHandler.List)
		r.Get("/users/{id}", userHandler.GetByID)
		r.Post("/products", productHandler.CreateProduct)
		r.Post("/products/{id}/variants", productHandler.CreateProductVariant)
		r.Post("/images", mediaHandler.UploadImage)
		r.Delete("/images/{id}", mediaHandler.DeleteImage)
		r.Post("/categories", categoryHandler.CreateCategory)
		r.Put("/categories/{id}", categoryHandler.UpdateCategory)
		r.Delete("/categories/{id}", categoryHandler.DeleteCategory)
		r.Get("/orders", orderHandler.ListOrders)
		r.Post("/orders", orderHandler.CreateOrder)
		r.Get("/orders/{id}", orderHandler.GetOrderByID)
		r.Put("/orders/{id}", orderHandler.UpdateOrder)
		r.Patch("/orders/{id}/status", orderHandler.UpdateOrderStatus)
		r.Delete("/orders/{id}", orderHandler.DeleteOrder)
		r.Delete("/orders/{id}/items/{itemId}", orderHandler.DeleteOrderItem)
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
	config  config
	db      *pgxpool.Pool
	storage *s3store.Store
}

type config struct {
	addr string
	db   dbConfig
	s3   s3store.Config
}

type dbConfig struct {
	dsn string
}
