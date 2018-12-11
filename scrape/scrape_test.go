package scrape

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vary/prom_rest_exporter/spec"
)

func TestScrape(t *testing.T) {
	ex, _ := spec.ReadSpecFromYamlFile("scrape_test_spec.yml")
	vals, err := ScrapeEndpoint(ex.Endpoints[0])

	assert.Nil(t, err)

	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "user_count", vals[0].Name)
	assert.Equal(t, 3, vals[0].value)

	assert.Equal(t, "user_count_total", vals[1].Name)
	assert.Equal(t, 12.5, vals[1].value)
}
