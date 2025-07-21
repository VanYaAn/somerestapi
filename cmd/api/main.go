package main

import (
	"context"
	"fmt"
	"os"
	"restapi/internal/config"
	"restapi/internal/handlers"
	"restapi/internal/logger"
	"restapi/internal/middleware"
	"restapi/internal/service"
	"restapi/internal/storage"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var JWTKey = os.Getenv("JWT_KEY")

func main() {
	JWTKey = string(JWTKey)
	if JWTKey == "" {
		fmt.Println("JWT_KEY environment variable is not set")
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Errorf("error loading config: %v", err)
	}
	fmt.Printf("Config loaded\n")

	logger, err := logger.NewLogger(cfg.Logger.Level)
	if err != nil {
		fmt.Errorf("error creating logger: %v", err)
	}
	fmt.Printf("Logger created with level: %s\n", logger.Level())

	if err != nil {
		logger.Errorf("error connecting to database: %v", err)
		return
	}
	storage := storage.New(context.Background(), cfg.Database.Port, cfg.Database.Username, cfg.Database.Host, cfg.Database.Name, cfg.Database.Password)
	if storage == nil {
		logger.Errorf("error creating storage: %v", err)
		return
	}
	defer storage.Database.Close()

	svc := service.NewService(cfg.Server.Port, cfg.Server.Host, logger, storage)

	r := mux.NewRouter()

	h := handlers.NewHandler(svc)
	//не нужен jwt token
	public := r.PathPrefix("/api/v1").Subrouter()
	public.HandleFunc("/register", h.RegisterHandler).Methods("POST")
	public.HandleFunc("/login", h.LoginHandler).Methods("POST")
	public.HandleFunc("/ads", h.GetAdsHandler).Methods("GET")

	//нужен jwt token
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware(logger, JWTKey))
	protected.HandleFunc("/ads", h.CreateAdHandler).Methods("POST")

	if err := svc.ListenAndServe(r); err != nil {
		logger.Errorf("error starting server: %v", err)
		os.Exit(1)
	}
}
