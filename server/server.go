package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
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
	printMetrics(w, vals)
}

func printMetrics(w io.Writer, vals []scrape.MetricValue) {
	for _, val := range vals {
		if val.Description != "" {
			fmt.Fprintf(w, "# HELP %s %s\n", val.Name, val.Description)
		}
		if val.Type != "" {
			fmt.Fprintf(w, "# TYPE %s %s\n", val.Name, val.Type)
		}
		fmt.Fprintf(w, "%s %s\n\n", val.Name, val.FormatVal())
	}
}
