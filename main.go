package main

import (
	"fmt"
	"vary/prom_rest_exporter/scrape"
	"vary/prom_rest_exporter/spec"
)

func main() {

	eps, _ := spec.ReadSpecFromYamlFile("sample.yml")

	vals, _ := scrape.ScrapeEndpoint(eps.Endpoints[0])
	for _, val := range vals {
		if val.Description != "" {
			fmt.Printf("# HELP %s %s\n", val.Name, val.Description)
		}
		if val.Type != "" {
			fmt.Printf("# TYPE %s %s\n", val.Name, val.Type)
		}
		fmt.Printf("%s %s\n\n", val.Name, val.FormatVal())
	}
}
