package server

import (
	"github.com/sandro-h/prom_rest_exporter/spec"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
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
	time.Sleep(100 * time.Millisecond)
}

func TestRequestMetrics(t *testing.T) {
	resp, err := tryFetch("http://localhost:9011/metrics", 3)
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

func tryFetch(url string, retries int) (string, error) {
	resp, err := fetch(url)
	for i := 0; i < retries && err != nil; i++ {
		time.Sleep(500 * time.Millisecond)
		resp, err = fetch(url)
	}
	return resp, err
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
