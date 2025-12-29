package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"csd-pilote/backend/modules/platform/config"
	"csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/database"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
)

// Server represents the csd-pilote server
type Server struct {
	cfg           *config.Config
	httpServer    *http.Server
	csdCoreClient *csdcore.Client
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	// Connect to database
	_, err := database.Connect(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Connected to database")

	// Create csd-core client
	csdCoreClient := csdcore.NewClient(&cfg.CSDCore)
	log.Printf("CSD-Core client configured: %s", cfg.CSDCore.URL)

	// Create GraphQL handler
	graphqlHandler := graphql.NewHandler(csdCoreClient)

	// Setup routes
	mux := http.NewServeMux()

	// API base path
	const apiBasePath = "/pilote/api/latest"

	// Health check
	mux.HandleFunc(apiBasePath+"/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
	})

	// GraphQL endpoint
	mux.Handle(apiBasePath+"/query", graphqlHandler)

	// Apply middleware
	handler := corsMiddleware(cfg.CORS)(middleware.AuthMiddleware(mux))

	server := &Server{
		cfg:           cfg,
		csdCoreClient: csdCoreClient,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
			Handler: handler,
		},
	}

	return server, nil
}

// Start starts the server
func (s *Server) Start() error {
	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	database.Close()
	log.Println("Server stopped")
	return nil
}

// corsMiddleware returns a CORS middleware
func corsMiddleware(cfg config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range cfg.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
