package scrape

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sandro-h/prom_rest_exporter/spec"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestScrape(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets, false)

	assert.Equal(t,
		`# HELP user_count Number of users
# TYPE user_count gauge
user_count 3

# HELP user_id User ids
# TYPE user_id gauge
user_id{last_name="Bluth"} 1
user_id{last_name="Weaver"} 2
user_id{last_name="Wong"} 3

`,
		printMetrics(metrics))
}

func TestScrapeDefaultLabel(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_default_lbl_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets, false)

	assert.Equal(t,
		`user_id{val_index="0"} 1
user_id{val_index="1"} 2
user_id{val_index="2"} 3

`,
		printMetrics(metrics))
}

func TestScrapeMultiLabels(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_multi_lbl_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets, false)

	assert.Equal(t,
		`user_id{first_name="George",last_name="Bluth"} 1
user_id{first_name="Janet",last_name="Weaver"} 2
user_id{first_name="Emma",last_name="Wong"} 3

`,
		printMetrics(metrics))
}

func TestScrapeFixedLabel(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_fixed_lbl_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets, false)

	assert.Equal(t,
		`# HELP user_count Number of users
# TYPE user_count gauge
user_count{foobar="hello"} 3

user_id{foobar="world",val_index="0"} 1
user_id{foobar="world",val_index="1"} 2
user_id{foobar="world",val_index="2"} 3

`,
		printMetrics(metrics))
}

func TestScrapeNotFoundValSkipped(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_no_val_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets, false)

	assert.Equal(t,
		`# HELP user_count2 Number of users
# TYPE user_count2 gauge
user_count2 3

`,
		printMetrics(metrics))
}

func TestScrapeNotFoundLabelSkipped(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_no_label_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets, false)

	assert.Equal(t,
		`# HELP user_id User ids
# TYPE user_id gauge
user_id{last_name="Bluth"} 1
user_id{last_name="Weaver"} 2
user_id{last_name="Wong"} 3

`,
		printMetrics(metrics))
}

func TestScrapeFetchErrorSkipped(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_fetch_error_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets, false)

	assert.Equal(t,
		`# HELP user_count2 Number of users
# TYPE user_count2 gauge
user_count2 3

`,
		printMetrics(metrics))
}

func TestScrapeBasicAuth(t *testing.T) {
	srv := StartTestRestServer(19011)
	defer srv.Stop()

	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_basic_auth_spec.yml")
	ScrapeTargets(spec.Endpoints[0].Targets, false)

	assert.Equal(t, 1, len(srv.ReceivedReqs))
	user, pwd, ok := srv.ReceivedReqs[0].BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "user123", user)
	assert.Equal(t, "pass123", pwd)
}

func TestScrapeHeaders(t *testing.T) {
	srv := StartTestRestServer(19011)
	defer srv.Stop()

	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_headers_spec.yml")
	ScrapeTargets(spec.Endpoints[0].Targets, false)

	assert.Equal(t, 1, len(srv.ReceivedReqs))
	assert.Equal(t, "CustomValue1", srv.ReceivedReqs[0].Header.Get("CustomHeader1"))
	assert.Equal(t, "CustomValue2", srv.ReceivedReqs[0].Header.Get("CustomHeader2"))
}

func TestMetaMetrics(t *testing.T) {
	fixedNow := time.Unix(1545391515, 0)
	getNow = func() time.Time {
		return fixedNow
	}
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets, true)

	assert.Equal(t,
		`# HELP prom_rest_exp_last_exec_time Unix timestamp of last execution
# TYPE prom_rest_exp_last_exec_time gauge
prom_rest_exp_last_exec_time 1545391515

# HELP prom_rest_exp_metric_fails Number of failures during metrics collection
# TYPE prom_rest_exp_metric_fails gauge
prom_rest_exp_metric_fails{url="file://testdata/scrape_test_data.json"} 0

# HELP prom_rest_exp_metrics_count Number of metrics returned in this call (not including same metric with multiple values)
# TYPE prom_rest_exp_metrics_count gauge
prom_rest_exp_metrics_count 2

# HELP prom_rest_exp_response_time Response time from REST endpoint
# TYPE prom_rest_exp_response_time gauge
prom_rest_exp_response_time{url="file://testdata/scrape_test_data.json"} 0

# HELP prom_rest_exp_values_count Number of values returned, including metric with multiple values
# TYPE prom_rest_exp_values_count gauge
prom_rest_exp_values_count 4

# HELP user_count Number of users
# TYPE user_count gauge
user_count 3

# HELP user_id User ids
# TYPE user_id gauge
user_id{last_name="Bluth"} 1
user_id{last_name="Weaver"} 2
user_id{last_name="Wong"} 3

`,
		printMetrics(metrics))
}

type ByMetricName []MetricInstance

func (a ByMetricName) Len() int           { return len(a) }
func (a ByMetricName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByMetricName) Less(i, j int) bool { return strings.Compare(a[i].Name, a[j].Name) < 0 }

func printMetrics(metrics []MetricInstance) string {
	var b bytes.Buffer
	sort.Sort(ByMetricName(metrics))
	for _, m := range metrics {
		m.PrintSortedLabels(&b)
	}
	return b.String()
}

type TestRestServer struct {
	srv          *http.Server
	ReceivedReqs []*http.Request
}

func StartTestRestServer(port int) *TestRestServer {
	srv := TestRestServer{}
	srv.ReceivedReqs = make([]*http.Request, 0)

	router := mux.NewRouter()
	router.HandleFunc("/test", srv.GetTestData).Methods("GET")

	srv.srv = &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("localhost:%d", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go srv.srv.ListenAndServe()
	return &srv
}

func (srv *TestRestServer) Stop() {
	srv.srv.Shutdown(nil)
}

func (srv *TestRestServer) GetTestData(w http.ResponseWriter, r *http.Request) {
	srv.ReceivedReqs = append(srv.ReceivedReqs, r)
}
