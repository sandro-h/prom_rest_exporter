package scrape

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"vary/prom_rest_exporter/spec"
)

type MetricInstance struct {
	values []MetricValue
	*spec.MetricSpec
}

type MetricValue struct {
	value     interface{} // float64 or int
	labelVals map[string]string
}

func (val *MetricInstance) Print(w io.Writer) {
	if val.Description != "" {
		fmt.Fprintf(w, "# HELP %s %s\n", val.Name, val.Description)
	}
	if val.Type != "" {
		fmt.Fprintf(w, "# TYPE %s %s\n", val.Name, val.Type)
	}

	for i, vi := range val.values {
		fqn := val.Name
		if len(vi.labelVals) > 0 {
			lbls := ""
			for n, v := range vi.labelVals {
				if lbls != "" {
					lbls += ","
				}
				lbls += n + "=\"" + v + "\""
			}
			fqn += "{" + lbls + "}"
		} else if len(val.values) > 1 {
			// If there is more than 1 value for the metric, but no labels
			// to distinguish them, add a label with the index.
			fqn += fmt.Sprintf("{val_index=\"%d\"}", i)
		}
		fmt.Fprintf(w, "%s %s\n", fqn, vi.formatVal())
	}
	fmt.Fprintf(w, "\n")
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

func ScrapeTargets(ts []*spec.TargetSpec) ([]MetricInstance, error) {
	allValues := make([]MetricInstance, 0)
	for _, t := range ts {
		values, _ := ScrapeTarget(t)
		allValues = append(allValues, values...)
	}
	return allValues, nil
}

func ScrapeTarget(t *spec.TargetSpec) ([]MetricInstance, error) {
	input, _ := fetch(t.Url)
	values := make([]MetricInstance, 0)
	// fmt.Printf("input:%s\n", input)
	for _, m := range t.Metrics {
		results, err := m.JqInst.ProcessInput(input)
		if err != nil {
			fmt.Printf("Process error: %s", err)
			// TODO
		}

		valInsts := make([]MetricValue, 0)
		for _, res := range results {
			realRes := res
			labelVals := make(map[string]string)
			if m.ValJqInst != nil {
				// TODO handle error
				subResults, _ := m.ValJqInst.ProcessInputJv(res)
				if len(subResults) > 0 {
					realRes = subResults[0]
				} else {
					realRes = nil
				}
			}

			for _, l := range m.Labels {

				// TODO handle error
				lblResults, _ := l.JqInst.ProcessInputJv(res)
				if len(lblResults) > 0 && lblResults[0].IsString() {
					labelVals[l.Name] = lblResults[0].ToString()
				}
			}

			if realRes != nil && realRes.IsNumber() {
				valInsts = append(valInsts, MetricValue{realRes.ToNumber(), labelVals})
			}
		}
		val := MetricInstance{valInsts, m}
		values = append(values, val)

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
