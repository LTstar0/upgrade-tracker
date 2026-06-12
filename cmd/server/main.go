package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"upgrade-tracker/internal/config"
	"upgrade-tracker/internal/db"
	"upgrade-tracker/internal/handler"
	"upgrade-tracker/internal/repo"
)

func main() {
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.New(cfg.Database)
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer database.Close()
	log.Println("✅ MySQL connected")

	clientRepo := repo.NewClientRepo(database)
	upgradeRepo := repo.NewUpgradeRepo(database)
	imageRepo := repo.NewImageRepo(database)

	mux := http.NewServeMux()

	// API routes
	h := handler.New(clientRepo, upgradeRepo, imageRepo)
	h.RegisterRoutes(mux)

	// Static frontend — serve index.html for all non-API paths
	frontendDir := "frontend"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		path := filepath.Join(frontendDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
			return
		}
		http.ServeFile(w, r, path)
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("🚀 Server running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
