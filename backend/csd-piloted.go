package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"csd-pilote/backend/modules/platform/config"
	"csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/server"

	// Import modules to register their GraphQL operations
	_ "csd-pilote/backend/modules/pilot/clusters"
	_ "csd-pilote/backend/modules/pilot/containers"
	_ "csd-pilote/backend/modules/pilot/dashboard"
	_ "csd-pilote/backend/modules/pilot/hypervisors"
	_ "csd-pilote/backend/modules/pilot/security"

	// Kubernetes resources
	_ "csd-pilote/backend/modules/pilot/kubernetes/deployments"
	_ "csd-pilote/backend/modules/pilot/kubernetes/namespaces"
	_ "csd-pilote/backend/modules/pilot/kubernetes/pods"
	_ "csd-pilote/backend/modules/pilot/kubernetes/services"

	// Libvirt resources
	_ "csd-pilote/backend/modules/pilot/libvirt/domains"
	_ "csd-pilote/backend/modules/pilot/libvirt/networks"
	_ "csd-pilote/backend/modules/pilot/libvirt/storage"
)

var Version = "1.0.0"

func main() {
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded")

	// Register with csd-core
	if cfg.CSDCore.ServiceToken != "" {
		log.Printf("Registering service with csd-core at %s%s...", cfg.CSDCore.URL, cfg.CSDCore.GraphQLEndpoint)
		if err := registerWithCore(cfg); err != nil {
			log.Printf("Warning: Failed to register with csd-core: %v", err)
		} else {
			log.Printf("Successfully registered as 'csd-pilote' with csd-core")
		}
	} else {
		log.Printf("Warning: No service-token configured, skipping csd-core registration")
	}

	// Create and start server
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Printf("CSD-Pilote server starting...")
	log.Printf("  Health:   http://%s:%s/pilote/api/latest/health", cfg.Server.Host, cfg.Server.Port)
	log.Printf("  GraphQL:  http://%s:%s/pilote/api/latest/query", cfg.Server.Host, cfg.Server.Port)
	log.Printf("  CSD-Core: %s", cfg.CSDCore.URL)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func registerWithCore(cfg *config.Config) error {
	client := csdcore.NewClient(&cfg.CSDCore)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serviceURL := fmt.Sprintf("http://%s:%s", cfg.Server.Host, cfg.Server.Port)
	if cfg.Server.Host == "0.0.0.0" {
		serviceURL = fmt.Sprintf("http://localhost:%s", cfg.Server.Port)
	}

	reg := &csdcore.ServiceRegistration{
		Name:        "CSD Pilote",
		Slug:        "csd-pilote",
		Version:     Version,
		BaseURL:     serviceURL,
		CallbackURL: serviceURL + "/pilote/api/latest/query",
		Description: "Infrastructure management application for Kubernetes and Libvirt",
		// Frontend integration (Module Federation)
		FrontendURL:     cfg.Frontend.URL,
		RemoteEntryPath: cfg.Frontend.RemoteEntryPath,
		RoutePath:       cfg.Frontend.RoutePath,
		ExposedModules: map[string]string{
			"./Routes":       "./src/Routes.tsx",
			"./Translations": "./src/translations/generated/index.ts",
		},
	}

	return client.RegisterService(ctx, cfg.CSDCore.ServiceToken, reg)
}
