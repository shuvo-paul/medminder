package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
)

type healthOutput struct {
	Body struct {
		Status    string `json:"status" doc:"Service status"`
		Timestamp string `json:"timestamp" doc:"Current server time in RFC3339 format"`
	}
}

func registerHealthRoute(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/api/healthz",
		Summary:     "Health check",
		Tags:        []string{"system"},
	}, func(_ context.Context, _ *struct{}) (*healthOutput, error) {
		resp := &healthOutput{}
		resp.Body.Status = "ok"
		resp.Body.Timestamp = time.Now().UTC().Format(time.RFC3339)
		return resp, nil
	})
}

func registerOpenAPIRoute(router *chi.Mux, api huma.API) {
	router.Get("/api/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/openapi+json")
		json.NewEncoder(w).Encode(api.OpenAPI()) //nolint:errcheck
	})
}
