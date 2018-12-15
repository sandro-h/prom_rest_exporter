package server

import (
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"vary/prom_rest_exporter/scrape"
	"vary/prom_rest_exporter/spec"
)

type MetricServer struct {
	Endpoint *spec.EndpointSpec
	srv      *http.Server
}

func (srv *MetricServer) Start() {
	log.Infof("Starting metric endpoint at localhost:%d/metrics", srv.Endpoint.Port)
	router := mux.NewRouter()
	router.HandleFunc("/metrics", srv.GetMetrics).Methods("GET")

	srv.srv = &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("localhost:%d", srv.Endpoint.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.srv.ListenAndServe()
}

func (srv *MetricServer) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	vals, _ := scrape.ScrapeTargets(srv.Endpoint.Targets)
	for _, val := range vals {
		val.Print(w)
	}
}
