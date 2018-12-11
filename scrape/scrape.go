package scrape

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"vary/prom_rest_exporter/spec"
)

type MetricValue struct {
	value interface{} // float64 or int
	*spec.MetricSpec
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

func ScrapeEndpoint(e *spec.EndpointSpec) ([]MetricValue, error) {
	input, _ := fetch(e.Url)
	values := make([]MetricValue, 0)
	// fmt.Printf("input:%s\n", input)
	for _, m := range e.Metrics {
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

// Fetch calls the url and returns the response as a string
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
