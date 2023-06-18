package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dai65527/go-circuit-breaker-example/app/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	port          = "18080"
	errorRate int = 0
	latency   int = 0
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

// settingHandler エラー率ととレイテンシをセット
func settingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	r.ParseForm()
	erValue, err := strconv.Atoi(r.PostFormValue("errorrate"))
	if err == nil {
		errorRate = erValue
	}
	latencyValue, err := strconv.Atoi(r.PostFormValue("latency"))
	if err == nil {
		latency = latencyValue
	}
	log.Printf("setting updated (errorRate: %d%%, latency: %dmsec)", errorRate, latency)
}

// doHandler 何か処理を行うフリ風
func doHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(latency) * time.Millisecond)
	if rand.Intn(100) < errorRate {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error")
		return
	}
	fmt.Fprintf(w, "done")
}

func main() {
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/setting", settingHandler)
	http.Handle("/do", metrics.GenInstrumentChain("backend.do", doHandler))
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+port, nil)
}
