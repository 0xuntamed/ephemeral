package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var page = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Preview {{.EnvironmentID}}</title>
<style>body{font-family:system-ui;margin:0;background:#07110e;color:#e7f9f0;display:grid;min-height:100vh;place-items:center}.card{width:min(680px,calc(100% - 48px));background:#10221c;border:1px solid #2b4b40;border-radius:24px;padding:36px;box-shadow:0 30px 80px #0008}.eyebrow{color:#7be0b6;text-transform:uppercase;letter-spacing:.14em;font-weight:700;font-size:12px}h1{font-size:clamp(34px,7vw,64px);line-height:1;margin:18px 0}.pill{display:inline-flex;padding:8px 12px;border-radius:999px;background:#18372d;color:#9ff4cf}dl{display:grid;grid-template-columns:max-content 1fr;gap:12px 24px;margin-top:30px}dt{color:#90aa9f}dd{margin:0;font-family:ui-monospace,monospace;overflow-wrap:anywhere}</style></head>
<body><main class="card"><div class="eyebrow">Ephemeral preview</div><h1>It works.</h1><span class="pill">Healthy &middot; isolated workload</span><dl><dt>Environment</dt><dd>{{.EnvironmentID}}</dd><dt>Database schema</dt><dd>{{.DatabaseSchema}}</dd><dt>Image build</dt><dd>{{.BuildSHA}}</dd></dl></main></body></html>`))

type preview struct {
	EnvironmentID  string `json:"environment_id"`
	DatabaseSchema string `json:"database_schema"`
	BuildSHA       string `json:"build_sha"`
}

func main() {
	data := preview{
		EnvironmentID:  value("ENVIRONMENT_ID", "local-demo"),
		DatabaseSchema: value("DB_SCHEMA", "preview_local_demo"),
		BuildSHA:       value("BUILD_SHA", "development"),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Execute(writer, data); err != nil {
			log.Printf("render page: %v", err)
		}
	})
	mux.HandleFunc("GET /healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]string{"status": "ok", "environment_id": data.EnvironmentID})
	})
	server := &http.Server{
		Addr: ":" + value("PORT", "8080"), Handler: mux,
		ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second,
	}
	fmt.Printf("demo service listening on %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func value(key, fallback string) string {
	if current := os.Getenv(key); current != "" {
		return current
	}
	return fallback
}
