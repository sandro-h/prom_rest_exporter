package scrape

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"vary/prom_rest_exporter/spec"
)

func TestScrape(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets)

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
	metrics := ScrapeTargets(spec.Endpoints[0].Targets)

	assert.Equal(t,
		`user_id{val_index="0"} 1
user_id{val_index="1"} 2
user_id{val_index="2"} 3

`,
		printMetrics(metrics))
}

func TestScrapeMultiLabels(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_multi_lbl_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets)

	assert.Equal(t,
		`user_id{first_name="George",last_name="Bluth"} 1
user_id{first_name="Janet",last_name="Weaver"} 2
user_id{first_name="Emma",last_name="Wong"} 3

`,
		printMetrics(metrics))
}

func TestScrapeFixedLabel(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_fixed_lbl_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets)

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
	metrics := ScrapeTargets(spec.Endpoints[0].Targets)

	assert.Equal(t,
		`# HELP user_count2 Number of users
# TYPE user_count2 gauge
user_count2 3

`,
		printMetrics(metrics))
}

func TestScrapeNotFoundLabelSkipped(t *testing.T) {
	spec, _ := spec.ReadSpecFromYamlFile("testdata/scrape_test_no_label_spec.yml")
	metrics := ScrapeTargets(spec.Endpoints[0].Targets)

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
	metrics := ScrapeTargets(spec.Endpoints[0].Targets)

	assert.Equal(t,
		`# HELP user_count2 Number of users
# TYPE user_count2 gauge
user_count2 3

`,
		printMetrics(metrics))
}

func printMetrics(metrics []MetricInstance) string {
	var b bytes.Buffer
	for _, m := range metrics {
		m.PrintSortedLabels(&b)
	}
	return b.String()
}
