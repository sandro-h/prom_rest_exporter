package main

import (
	"bufio"
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"vary/prom_rest_exporter/server"
	"vary/prom_rest_exporter/spec"
)

var debug = flag.Bool("debug", false, "Enables detailed debug logging")
var trace = flag.Bool("trace", false, "Enables most detailed trace logging. Overrides the debug flag.")
var config = flag.String("config", "prom_rest_exporter.yml", "Set path to config yaml file. Default: prom_rest_exporter.yml")

func main() {
	flag.Parse()
	logFile := initLogging()
	defer logFile.Close()

	log.Infof("Starting prom_rest_exporter with config file %s", *config)

	spec, err := spec.ReadSpecFromYamlFile(*config)
	if err != nil {
		log.Errorf("Error reading %s: %s", *config, err)
		panic(err)
	}
	for _, ep := range spec.Endpoints {
		srv := server.MetricServer{Endpoint: ep}
		go srv.Start()
	}

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func initLogging() *os.File {
	file, err := os.OpenFile("prom_rest_exporter.log", os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	log.SetOutput(file)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	if *trace {
		log.SetReportCaller(true)
		log.SetLevel(log.TraceLevel)
	}
	return file
}
