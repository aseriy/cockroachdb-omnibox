package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"omnibox-visualize/internal/config"
	"omnibox-visualize/internal/tail"
)

const configFile = "visualize.yml"

//go:embed all:ui/dist
var ui embed.FS

func main() {
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatal(err)
	}

	path := make(map[string]string, len(cfg.Lanes))
	for _, lane := range cfg.Lanes {
		path[lane.Name] = lane.Path
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cfg); err != nil {
			log.Print(err)
		}
	})
	mux.HandleFunc("GET /api/stream/{lane}", func(w http.ResponseWriter, r *http.Request) {
		stream(w, r, path, cfg.BufferSize)
	})
	dist, err := fs.Sub(ui, "ui/dist")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /", http.FileServerFS(dist))

	log.Printf("listening on %s", cfg.Listen)
	log.Fatal(http.ListenAndServe(cfg.Listen, mux))
}

// stream tails one lane's log file for as long as the client stays connected,
// writing every record as an SSE data frame.
func stream(w http.ResponseWriter, r *http.Request, path map[string]string, seed int) {
	name := r.PathValue("lane")
	file, ok := path[name]
	if !ok {
		http.NotFound(w, r)
		return
	}
	log.Printf("lane %s following %s", name, file)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	control := http.NewResponseController(w)
	started := false
	err := tail.Tail(r.Context(), file, seed, func(line tail.Line) error {
		payload, err := json.Marshal(line)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
			return err
		}
		started = true
		return control.Flush()
	})

	// Nothing has reached the wire if the file could not be opened, so the
	// status line is still ours to write.
	if err != nil && !started {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
