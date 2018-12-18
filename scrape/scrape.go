package scrape

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"vary/prom_rest_exporter/jq"
	"vary/prom_rest_exporter/spec"
)

func ScrapeTargets(ts []*spec.TargetSpec) []MetricInstance {
	allValues := make([]MetricInstance, 0)
	for _, t := range ts {
		values, err := ScrapeTarget(t)
		if err != nil {
			log.Errorf("Error scraping target %s: %s", t.URL, err)
		} else {
			allValues = append(allValues, values...)
		}
	}
	return allValues
}

func ScrapeTarget(t *spec.TargetSpec) ([]MetricInstance, error) {
	log.Debugf("Scraping target %s", t.URL)
	input, err := fetch(t.URL)
	if err != nil {
		return nil, err
	}
	log.Tracef("Data from %s: %s", t.URL, input)

	metrics := make([]MetricInstance, 0)
	for _, m := range t.Metrics {
		results, err := m.JqInst.ProcessInput(input)
		if err != nil {
			log.Errorf("Error processing input of %s for metric %s: %s", t.URL, m.Name, err)
			continue
		}

		metricVals := make([]MetricValue, 0)
		for _, res := range results {
			val := getValue(m, res)

			if val == nil {
				log.Errorf("Error processing input of %s for metric %s: no valid value found", t.URL, m.Name)
			} else {
				labels := getLabels(m, res)
				metricVals = append(metricVals, MetricValue{val, labels})
			}
		}
		if len(metricVals) > 0 {
			val := MetricInstance{metricVals, m}
			metrics = append(metrics, val)
		}
	}

	return metrics, nil
}

func getValue(m *spec.MetricSpec, res *jq.Jv) interface{} {
	val := res
	if m.ValJqInst != nil {
		subResults, err := m.ValJqInst.ProcessInputJv(res)
		if err != nil {
			log.Errorf("Error getting value for metric %s: %s", m.Name, err)
		}
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
			lblResults, err := l.JqInst.ProcessInputJv(res)
			if err != nil {
				log.Errorf("Error getting label for metric %s: %s", m.Name, err)
			} else if len(lblResults) > 0 && lblResults[0].IsString() {
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
