package scrape

import (
	"github.com/sandro-h/prom_rest_exporter/jq"
	"github.com/sandro-h/prom_rest_exporter/spec"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var getNow = func() time.Time {
	return time.Now()
}

// ScrapeTargets calls the REST endpoints in the passed targets and extracts metrics
func ScrapeTargets(ts []*spec.TargetSpec, inclMetaMetrics bool) []MetricInstance {
	allMetrics := make([]MetricInstance, 0)

	var metas map[string]*MetricInstance
	var metasPtr *map[string]*MetricInstance
	if inclMetaMetrics {
		metas = make(map[string]*MetricInstance)
		metasPtr = &metas
	}

	for _, t := range ts {
		metrics, err := scrapeTarget(t, metasPtr)
		if err != nil {
			log.Errorf("Error scraping target %s: %s", t.URL, err)
		} else {
			allMetrics = append(allMetrics, *metrics...)
		}
	}

	if inclMetaMetrics {
		computeOverallMetaMetrics(metasPtr, &allMetrics)
		for _, m := range metas {
			allMetrics = append(allMetrics, *m)
		}
	}

	return allMetrics
}

func scrapeTarget(t *spec.TargetSpec, metas *map[string]*MetricInstance) (*[]MetricInstance, error) {
	log.Debugf("Scraping target %s", t.URL)
	tm := getNow()
	restResponse, err := fetch(t.URL, t.User, t.Password, &t.Headers)
	fetchDuration := getNow().Sub(tm)
	if err != nil {
		return nil, err
	}
	log.Tracef("Data from %s: %s", t.URL, restResponse)

	metrics, metricFails := extractMetrics(t, &restResponse)

	if metas != nil {
		computeTargetMetaMetrics(metas, t.URL, fetchDuration, metricFails)
	}

	return metrics, nil
}

func extractMetrics(t *spec.TargetSpec, restResponse *string) (*[]MetricInstance, int) {
	metrics := make([]MetricInstance, 0)
	metricFails := 0
	for _, m := range t.Metrics {
		results, err := m.JqInst.ProcessInput(*restResponse)
		if err != nil {
			log.Errorf("Error processing input of %s for metric %s: %s", t.URL, m.Name, err)
			metricFails++
		} else {
			metricVals := make([]MetricValue, 0)
			for _, res := range results {
				val := getValue(m, res)

				if val == nil {
					log.Errorf("Error processing input of %s for metric %s: no valid value found", t.URL, m.Name)
					metricFails++
				} else {
					labels := getLabels(m, res)
					metricVals = append(metricVals, MetricValue{val, labels})
				}
			}
			if len(metricVals) > 0 {
				val := MetricInstance{metricVals, m}
				metrics = append(metrics, val)
			}
			freeResults(results)
		}
	}

	return &metrics, metricFails
}

// Does not consume res. Returns int, float64, or nil
func getValue(m *spec.MetricSpec, res *jq.Jv) interface{} {
	val := res
	if m.ValJqInst != nil {
		subResults, err := m.ValJqInst.ProcessInputJv(res)
		defer freeResults(subResults)
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

func freeResults(res []*jq.Jv) {
	for _, r := range res {
		r.Free()
	}
}

// Does not consume res
func getLabels(m *spec.MetricSpec, res *jq.Jv) map[string]string {
	labels := make(map[string]string)
	for _, l := range m.Labels {
		if l.FixedValue != "" {
			labels[l.Name] = l.FixedValue
		} else {
			lblResults, err := l.JqInst.ProcessInputJv(res)
			if err != nil {
				log.Errorf("Error getting label for metric %s: %s", m.Name, err)
			} else {
				if len(lblResults) > 0 && lblResults[0].IsString() {
					labels[l.Name] = lblResults[0].ToString()
				}
				freeResults(lblResults)
			}
		}
	}
	return labels
}

func computeTargetMetaMetrics(metas *map[string]*MetricInstance,
	fetchURL string, fetchDuration time.Duration,
	metricFails int) {
	addMetaMetric(metas,
		NewWithIntValue("prom_rest_exp_response_time", int(fetchDuration/time.Millisecond),
			"Response time from REST endpoint",
			"gauge",
			"url",
			fetchURL))

	addMetaMetric(metas,
		NewWithIntValue("prom_rest_exp_metric_fails", metricFails,
			"Number of failures during metrics collection",
			"gauge",
			"url",
			fetchURL))
}

func computeOverallMetaMetrics(metas *map[string]*MetricInstance, metrics *[]MetricInstance) {
	addMetaMetric(metas,
		NewWithIntValue("prom_rest_exp_metrics_count", len(*metrics),
			"Number of metrics returned in this call (not including same metric with multiple values)",
			"gauge",
			"",
			""))

	totalValues := 0
	for _, m := range *metrics {
		totalValues += len(m.values)
	}

	addMetaMetric(metas,
		NewWithIntValue("prom_rest_exp_values_count", totalValues,
			"Number of values returned, including metric with multiple values",
			"gauge",
			"",
			""))

	addMetaMetric(metas,
		NewWithIntValue("prom_rest_exp_last_exec_time", int(getNow().Unix()),
			"Unix timestamp of last execution",
			"gauge",
			"",
			""))
}

func addMetaMetric(metas *map[string]*MetricInstance, m MetricInstance) {
	exi, ok := (*metas)[m.Name]
	if ok {
		exi.values = append(exi.values, m.values...)
	} else {
		(*metas)[m.Name] = &m
	}
}

// Fetch makes a request to the url and returns the response as a string
func fetch(url string, user string, pwd string, headers *map[string]string) (string, error) {
	if strings.HasPrefix(url, "file://") {
		data, err := ioutil.ReadFile(url[7:])
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if user != "" && pwd != "" {
		req.SetBasicAuth(user, pwd)
	}
	if headers != nil {
		for k, v := range *headers {
			req.Header.Set(k, v)
		}
	}

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
