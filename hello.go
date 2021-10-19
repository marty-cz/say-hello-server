package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "http_requests_total",
			Help:      "count http responses",
		},
		[]string{"code", "method"},
	)

	responseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "http_response_total",
			Help:      "count http requests",
		},
		[]string{"status", "code"},
	)

	hello = map[string]string{
		"en": "Hello",
		"es": "Hola",
		"de": "Hallo",
		"ch": "你好",
		"cs": "Ahoj",
	}

	port           = getEnv("PORT", "8080")
	urlWithLangFmt = "http://0.0.0.0:" + port + "/%s"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func handler(w http.ResponseWriter, r *http.Request) {
	lang := strings.TrimPrefix(r.URL.RequestURI(), "/")
	greeting, ok := hello[lang]
	if !ok {
		fmt.Printf("unknown language: %s\n", lang)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, greeting)
}

func selfPing() {
	oscillationFactor := func() float64 {
		return 2 * rand.Float64()
	}

	var langs []string
	for key := range hello {
		langs = append(langs, key)
	}

	for {
		time.Sleep(time.Duration(oscillationFactor()) * time.Second)

		index := rand.Intn(len(langs) - 1)
		lang := langs[index]

		//log.Println("Ask for " + fmt.Sprintf(urlWithLangFmt, lang))
		resp, err := http.Get(fmt.Sprintf(urlWithLangFmt, lang))
		kind := "failure"
		if err != nil {
			log.Println("got am error", err)
			responseCounter.WithLabelValues(kind, "").Inc()
			continue
		}

		if resp.StatusCode == http.StatusOK {
			kind = "success"
		}

		io.Copy(os.Stdout, resp.Body)
		fmt.Println()
		resp.Body.Close()

		responseCounter.WithLabelValues(kind, resp.Status).Inc()
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	pr := prometheus.NewRegistry()
	pr.MustRegister(requestCounter, responseCounter)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(pr, promhttp.HandlerOpts{}))
	mux.Handle("/", promhttp.InstrumentHandlerCounter(requestCounter, http.HandlerFunc(handler)))

	log.Println("starting server on port ", port)

	go selfPing()

	log.Fatal(http.ListenAndServe(":"+port, mux))
}
