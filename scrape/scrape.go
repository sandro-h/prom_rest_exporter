package scrape

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"vary/prom_rest_exporter/spec"
)

type MetricValue struct {
	value interface{} // float64 or int
	*spec.MetricSpec
}

func (val *MetricValue) Print(w io.Writer) {
	if val.Description != "" {
		fmt.Fprintf(w, "# HELP %s %s\n", val.Name, val.Description)
	}
	if val.Type != "" {
		fmt.Fprintf(w, "# TYPE %s %s\n", val.Name, val.Type)
	}
	fmt.Fprintf(w, "%s %s\n\n", val.Name, val.formatVal())
}

func (mv *MetricValue) formatVal() string {
	switch v := mv.value.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	default:
		return "?"
	}
}

func ScrapeTargets(ts []*spec.TargetSpec) ([]MetricValue, error) {
	allValues := make([]MetricValue, 0)
	for _, t := range ts {
		values, _ := ScrapeTarget(t)
		allValues = append(allValues, values...)
	}
	return allValues, nil
}

func ScrapeTarget(t *spec.TargetSpec) ([]MetricValue, error) {
	input, _ := fetch(t.Url)
	values := make([]MetricValue, 0)
	// fmt.Printf("input:%s\n", input)
	for _, m := range t.Metrics {
		results, err := m.JqInst.ProcessInputFirstOnly(input)
		if err != nil {
			fmt.Printf("Process error: %s", err)
			// TODO
		}

		if len(results) > 0 && results[0].IsNumber() {
			val := MetricValue{results[0].ToNumber(), m}
			values = append(values, val)
		}
	}

	return values, nil
}

// Fetch makes a request to the url and returns the response as a string
func fetch(url string) (string, error) {
	if strings.HasPrefix(url, "file://") {
		data, err := ioutil.ReadFile(url[7:])
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

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
