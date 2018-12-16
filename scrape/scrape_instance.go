package scrape

import (
	"fmt"
	"io"
	"sort"
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
