// Sector Opportunity Analyzer - Go Implementation
// Single binary server with embedded static files
package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"sector-analyzer/api"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// Get port from environment or default to 8000
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Scores endpoints
		r.Get("/scores", api.GetScoresHandler)
		r.Get("/scores/summary", api.GetSummaryHandler)
		r.Get("/scores/{sector}", api.GetSectorScoreHandler)

		// Data endpoints
		r.Get("/data/sectors", api.GetSectorsHandler)
		r.Get("/data/quality", api.GetDataQualityHandler)

		// Cache endpoints
		r.Get("/cache/info", api.GetCacheInfoHandler)
		r.Post("/cache/clear", api.ClearCacheHandler)
	})

	// Health check
	r.Get("/health", api.HealthHandler)

	// Root endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if Accept header wants JSON
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "application/json") {
			api.RootHandler(w, r)
			return
		}
		// Otherwise serve the frontend
		serveStaticFile(w, r, "index.html")
	})

	// Serve static files for the frontend
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		serveStaticFile(w, r, path)
	})

	fmt.Printf("Starting Sector Opportunity Analyzer on port %s\n", port)
	fmt.Println("API endpoints:")
	fmt.Println("  GET  /health          - Health check")
	fmt.Println("  GET  /api/scores      - Get all sector scores")
	fmt.Println("  GET  /api/scores/summary - Get summary report")
	fmt.Println("  GET  /api/scores/{sector} - Get single sector score")
	fmt.Println("  GET  /api/data/sectors - List all sectors")
	fmt.Println("  GET  /api/cache/info  - Cache statistics")
	fmt.Println("  POST /api/cache/clear - Clear cache")

	log.Fatal(http.ListenAndServe(":"+port, r))
}

// serveStaticFile serves a file from the embedded static directory.
func serveStaticFile(w http.ResponseWriter, r *http.Request, path string) {
	// Get the static subdirectory
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Try to open the file
	file, err := staticFS.Open(path)
	if err != nil {
		// If file not found, serve index.html for SPA routing
		file, err = staticFS.Open("index.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		path = "index.html"
	}
	defer file.Close()

	// Get file info for content length
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If it's a directory, try to serve index.html from it
	if stat.IsDir() {
		file.Close()
		file, err = staticFS.Open(path + "/index.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		defer file.Close()
		stat, _ = file.Stat()
		path = path + "/index.html"
	}

	// Set content type based on file extension
	contentType := getContentType(path)
	w.Header().Set("Content-Type", contentType)

	// Read and write the file properly
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// getContentType returns the MIME type for a file path.
func getContentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".json"):
		return "application/json; charset=utf-8"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".ico"):
		return "image/x-icon"
	case strings.HasSuffix(path, ".woff"):
		return "font/woff"
	case strings.HasSuffix(path, ".woff2"):
		return "font/woff2"
	default:
		return "application/octet-stream"
	}
}
