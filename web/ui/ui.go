package ui

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:generate pnpm i
//go:generate pnpm run build
//go:embed all:dist/*
var uiFS embed.FS

func handler() http.Handler {
	sub, err := fs.Sub(uiFS, "dist")
	if err != nil {
		log.Fatalf("Failed to load subdirectory in UI handler: %s", err.Error())
	}

	return http.FileServer(http.FS(sub))
}

// ServeSPA serves the SPA and its assets
// Redirects all requests to the SPA's index.html except for /assets
func ServeSPA(w http.ResponseWriter, r *http.Request) {
	// Check if it starts with /assets
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		rawIndex, err := uiFS.ReadFile("dist/index.html")
		if err != nil {
			log.Fatalf("Failed to load index.html: %s", err.Error())
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(rawIndex)
		return
	}

	handler().ServeHTTP(w, r)
}
