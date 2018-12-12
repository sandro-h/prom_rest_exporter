package main

import (
	"bufio"
	"os"
	"vary/prom_rest_exporter/server"
	"vary/prom_rest_exporter/spec"
)

func main() {

	spec, _ := spec.ReadSpecFromYamlFile("prom_rest_exporter.yml")

	for _, ep := range spec.Endpoints {
		srv := server.MetricServer{Endpoint: ep}
		go srv.Start()
	}

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
