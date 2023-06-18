package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

var errorRate int
var latency int

// settingHandler エラー率ととレイテンシをセット
func settingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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
	if rand.Int63()%100 < int64(errorRate) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error")
		return
	}
	fmt.Fprintf(w, "done")
}

func main() {
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/do", doHandler)
	http.HandleFunc("/setting", settingHandler)
	http.ListenAndServe(":18080", nil)
}
