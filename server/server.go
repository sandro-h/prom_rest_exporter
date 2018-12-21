package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"vary/prom_rest_exporter/scrape"
	"vary/prom_rest_exporter/spec"
)

type MetricServer struct {
	Endpoint                *spec.EndpointSpec
	DefaultCacheTimeSeconds int
	srv                     *http.Server
	cache                   *cache.Cache
}

func (srv *MetricServer) Start() {
	host := "localhost"
	if srv.Endpoint.Host != "" {
		host = srv.Endpoint.Host
	}
	log.Infof("Starting metric endpoint at %s:%d/metrics", host, srv.Endpoint.Port)

	var ct time.Duration
	if srv.Endpoint.CacheTimeSeconds > 0 {
		ct = time.Duration(srv.Endpoint.CacheTimeSeconds)
	} else {
		ct = time.Duration(srv.DefaultCacheTimeSeconds)
	}
	log.Debugf("Using %ds cache time", ct)
	srv.cache = cache.New(ct*time.Second, 10*time.Minute)

	router := mux.NewRouter()
	router.HandleFunc("/metrics", srv.GetMetrics).Methods("GET")

	srv.srv = &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%d", host, srv.Endpoint.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.srv.ListenAndServe()
}

func (srv *MetricServer) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	var vals []scrape.MetricInstance
	cachedVals, found := srv.cache.Get("metrics")
	if found {
		vals = cachedVals.([]scrape.MetricInstance)
	} else {
		vals = scrape.ScrapeTargets(srv.Endpoint.Targets, srv.Endpoint.InclMetaMetrics)
		srv.cache.Set("metrics", vals, cache.DefaultExpiration)
	}

	for _, val := range vals {
		val.Print(w)
	}
}
