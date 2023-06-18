package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/dai65527/go-circuit-breaker-example/app/pkg/metrics"
	"github.com/mercari/go-circuitbreaker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	backendEndpoint = "http://localhost:18080"
	port            = "8080"

	cb = circuitbreaker.New(
		circuitbreaker.WithFailOnContextCancel(true),
		circuitbreaker.WithFailOnContextDeadline(true),
		circuitbreaker.WithHalfOpenMaxSuccesses(10),
		circuitbreaker.WithOpenTimeoutBackOff(backoff.NewExponentialBackOff()),
		circuitbreaker.WithOpenTimeout(10*time.Second),
		circuitbreaker.WithCounterResetInterval(10*time.Second),
		circuitbreaker.WithTripFunc(circuitbreaker.NewTripFuncFailureRate(10, 0.15)),
		circuitbreaker.WithOnStateChangeHookFn(func(from, to circuitbreaker.State) {
			log.Printf("state changed from %s to %s\n", from, to)
		}),
	)
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func do(ctx context.Context) error {
	resp, err := http.Get(backendEndpoint + "/do")
	if err != nil || resp.StatusCode != 200 {
		if err != nil {
			log.Print(err)
		}
		return errors.New("upstream error")
	}
	resp.Body.Close()
	return nil
}

func doWithCircuitBreaker(ctx context.Context) error {
	_, err := cb.Do(ctx, func() (interface{}, error) {
		err := do(ctx)
		return nil, err
	})

	if errors.Is(err, circuitbreaker.ErrOpen) {
		// 表示見やすいように40msecまつ
		time.Sleep(40 * time.Millisecond)
	}
	return err
}

// doHandler 何かをするやつ（backendにプロキシするだけ）
func doHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.URL.Query().Get("cb") == "1" {
		err = doWithCircuitBreaker(r.Context())
	} else {
		err = do(r.Context())
	}
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
