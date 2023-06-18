package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/dai65527/go-circuit-breaker-example/app/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	backendEndpoint = "http://localhost:18080"
	port            = "8080"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func do(ctx context.Context) error {
	resp, err := http.Get(backendEndpoint + "/do")
	if err != nil || resp.StatusCode != 200 {
		return errors.New("upstream error")
	}
	return nil
}

// doHandler 何かをするやつ（backendにプロキシするだけ）
func doHandler(w http.ResponseWriter, r *http.Request) {
	err := do(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error")
		return
	}
	fmt.Fprintf(w, "done")
}

func main() {
	if os.Getenv("BACKEND_ENDPOINT") != "" {
		backendEndpoint = os.Getenv("BACKEND_ENDPOINT")
	}
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	http.HandleFunc("/ping", pingHandler)
	http.Handle("/do", metrics.GenInstrumentChain("frontend.do", doHandler))
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+port, nil)
}
