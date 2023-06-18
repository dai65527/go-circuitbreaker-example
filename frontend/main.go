package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

var backendEndpoint = "http://localhost:18080"

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
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/do", doHandler)
	http.ListenAndServe(":8080", nil)
}
