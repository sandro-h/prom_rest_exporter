package server

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"vary/prom_rest_exporter/spec"
)

func TestMain(m *testing.M) {
	startup()
	retCode := m.Run()
	os.Exit(retCode)
}

func startup() {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/server_test_spec.yml")
	srv := MetricServer{Endpoint: spec.Endpoints[0]}
	go srv.Start()
}

func TestRequestMetrics(t *testing.T) {
	resp, err := fetch("http://localhost:9011/metrics")
	assert.Nil(t, err)
	assert.Equal(t,
		`# HELP user_count Number of users
# TYPE user_count gauge
user_count 3

# HELP user_count_total Total number of users
# TYPE user_count_total gauge
user_count_total 12.500000

`,
		resp)
}

func fetch(url string) (string, error) {
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
