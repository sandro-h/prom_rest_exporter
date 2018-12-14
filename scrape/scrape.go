package scrape

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"vary/prom_rest_exporter/jq"
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
	val.print(w, false)
}

func (val *MetricInstance) PrintSortedLabels(w io.Writer) {
	val.print(w, true)
}

func (m *MetricInstance) print(w io.Writer, sortLabels bool) {
	if m.Description != "" {
		fmt.Fprintf(w, "# HELP %s %s\n", m.Name, m.Description)
	}
	if m.Type != "" {
		fmt.Fprintf(w, "# TYPE %s %s\n", m.Name, m.Type)
	}

	for i, val := range m.values {
		// If there is more than 1 value for the metric, but no labels
		// to distinguish them, add a label with the index.
		needsValIndex := len(m.values) > 1 && (m.OnlyFixedLabels || len(val.labelVals) == 0)
		lbls := val.formatLabelString(i, sortLabels, needsValIndex)
		fmt.Fprintf(w, "%s%s %s\n", m.Name, lbls, val.formatVal())
	}
	fmt.Fprintf(w, "\n")
}

func (val *MetricValue) formatLabelString(valIndex int, sortLabels bool, addValIndex bool) string {
	lbls := ""
	if len(val.labelVals) > 0 {
		if sortLabels {
			var keys []string
			for k := range val.labelVals {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, n := range keys {
				lbls = concatLabel(lbls, n, val.labelVals[n])
			}
		} else {
			for n, v := range val.labelVals {
				lbls = concatLabel(lbls, n, v)
			}
		}
	}
	if addValIndex {
		lbls = concatLabel(lbls, "val_index", fmt.Sprintf("%d", valIndex))
	}

	if lbls == "" {
		return ""
	}
	return "{" + lbls + "}"
}

func concatLabel(lbls string, name string, val string) string {
	if lbls != "" {
		lbls += ","
	}
	return lbls + name + "=\"" + val + "\""
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
	metrics := make([]MetricInstance, 0)
	for _, m := range t.Metrics {
		results, err := m.JqInst.ProcessInput(input)
		if err != nil {
			fmt.Printf("Process error: %s", err)
			// TODO
		}

		metricVals := make([]MetricValue, 0)
		for _, res := range results {
			val := getValue(m, res)

			if val != nil {
				labels := getLabels(m, res)
				metricVals = append(metricVals, MetricValue{val, labels})
			}

		}
		val := MetricInstance{metricVals, m}
		metrics = append(metrics, val)

	}

	return metrics, nil
}

func getValue(m *spec.MetricSpec, res *jq.Jv) interface{} {
	val := res
	if m.ValJqInst != nil {
		// TODO handle error
		subResults, _ := m.ValJqInst.ProcessInputJv(res)
		if len(subResults) > 0 {
			val = subResults[0]
		} else {
			val = nil
		}
	}
	if val != nil && val.IsNumber() {
		return val.ToNumber()
	}
	return nil
}

func getLabels(m *spec.MetricSpec, res *jq.Jv) map[string]string {
	labels := make(map[string]string)
	for _, l := range m.Labels {
		if l.FixedValue != "" {
			labels[l.Name] = l.FixedValue
		} else {
			// TODO handle error
			lblResults, _ := l.JqInst.ProcessInputJv(res)
			if len(lblResults) > 0 && lblResults[0].IsString() {
				labels[l.Name] = lblResults[0].ToString()
			}
		}
	}
	return labels
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
