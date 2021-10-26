package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
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

func getLogLevelFromSysEnv() log.Level {
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}
	return logLevel
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(getLogLevelFromSysEnv())
}

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
		log.Errorf("Unknown language %s", lang)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, greeting)
}

func processResponse(resp *http.Response, err error) {
	defer resp.Body.Close()

	kind := "failure"
	if err != nil {
		log.Errorf("Got an error ", err)
		responseCounter.WithLabelValues(kind, "").Inc()
		return
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		kind = "success"
		log.Info(string(bodyBytes))
	} else {
		log.Warn(string(bodyBytes))
	}

	responseCounter.WithLabelValues(kind, resp.Status).Inc()
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

		index := rand.Intn(len(langs))
		lang := langs[index]

		log.Debugf("Ask for url %s", fmt.Sprintf(urlWithLangFmt, lang))
		resp, err := http.Get(fmt.Sprintf(urlWithLangFmt, lang))
		processResponse(resp, err)
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	pr := prometheus.NewRegistry()
	pr.MustRegister(requestCounter, responseCounter)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(pr, promhttp.HandlerOpts{}))
	mux.Handle("/", promhttp.InstrumentHandlerCounter(requestCounter, http.HandlerFunc(handler)))

	log.Infof("starting server on port %v", port)

	go selfPing()

	log.Fatal(http.ListenAndServe(":"+port, mux))
}
