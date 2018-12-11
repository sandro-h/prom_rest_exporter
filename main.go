package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"vary/prom_rest_exporter/jq"
)

type ExporterSpec struct {
	Endpoints []*EndpointSpec
}

type EndpointSpec struct {
	Url     string
	Metrics []*MetricSpec
}

type MetricSpec struct {
	Name        string
	Description string
	Type        string
	Selector    string
	jqInst      *jq.Jq
}

type MetricValue struct {
	metric *MetricSpec
	value  interface{} // float64 or int
}

func (mv *MetricValue) FormatVal() string {
	switch v := mv.value.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	default:
		return "?"
	}
}

func (es EndpointSpec) String() string {
	data, _ := yaml.Marshal(es)
	return string(data)
}

func (ex *ExporterSpec) ReadFromYamlFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &ex)
	if err != nil {
		return err
	}

	ex.compileMetrics()
	if err != nil {
		return err
	}

	return nil
}

func (ex *ExporterSpec) compileMetrics() error {
	for _, e := range ex.Endpoints {
		for _, m := range e.Metrics {
			m.jqInst = jq.New()
			err := m.jqInst.CompileProgram(m.Selector)
			if err != nil {
				fmt.Printf("Metric compile error: %s\n", err.Error())
				// Handle somehow, either fail fast or ignore and expulse metric
			}
		}
	}

	return nil
}

func scrapeEndpoint(e *EndpointSpec) ([]MetricValue, error) {
	input, _ := Fetch(e.Url)
	values := make([]MetricValue, 0)
	// fmt.Printf("input:%s\n", input)
	for _, m := range e.Metrics {
		results, err := m.jqInst.ProcessInputFirstOnly(input)
		if err != nil {
			fmt.Printf("Process error: %s", err)
			// TODO
		}

		if len(results) > 0 && results[0].IsNumber() {
			val := MetricValue{metric: m, value: results[0].ToNumber()}
			values = append(values, val)
		}
	}

	return values, nil
}

func main() {

	eps := ExporterSpec{}
	eps.ReadFromYamlFile("sample.yml")

	vals, _ := scrapeEndpoint(eps.Endpoints[0])
	for _, val := range vals {
		if val.metric.Description != "" {
			fmt.Printf("# HELP %s %s\n", val.metric.Name, val.metric.Description)
		}
		if val.metric.Type != "" {
			fmt.Printf("# TYPE %s %s\n", val.metric.Name, val.metric.Type)
		}
		fmt.Printf("%s %s\n\n", val.metric.Name, val.FormatVal())
	}

	// jqInst := jq.New()
	// defer jqInst.Close()

	// err := jqInst.CompileProgram("[.data[].last_name] | length")
	// if err != nil {
	// 	fmt.Printf("Compile error: %s", err)
	// 	return
	// }

	// input, _ := Fetch("https://reqres.in/api/users")

	// results, err := jqInst.ProcessInput(input)
	// if err != nil {
	// 	fmt.Printf("Process error: %s", err)
	// 	return
	// }

	// for _, r := range results {
	// 	r.PrettyPrint()
	// }
}

// Fetch calls the url and returns the response as a string
func Fetch(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
