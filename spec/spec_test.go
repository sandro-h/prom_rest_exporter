package spec

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadSpecFromYaml(t *testing.T) {
	ex, err := ReadSpecFromYamlFile("spec_test_spec.yml")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ex.Endpoints))

	assert.Equal(t, "https://reqres.in/api/users", ex.Endpoints[0].Url)
	assert.Equal(t, 2, len(ex.Endpoints[0].Metrics))
	assert.Equal(t, "user_count", ex.Endpoints[0].Metrics[0].Name)
	assert.Equal(t, "Number of users", ex.Endpoints[0].Metrics[0].Description)
	assert.Equal(t, "gauge", ex.Endpoints[0].Metrics[0].Type)
	assert.Equal(t, "[.data[].last_name] | length", ex.Endpoints[0].Metrics[0].Selector)

	assert.Equal(t, "user_count_total", ex.Endpoints[0].Metrics[1].Name)
	assert.Equal(t, "Total number of users", ex.Endpoints[0].Metrics[1].Description)
	assert.Equal(t, "gauge", ex.Endpoints[0].Metrics[1].Type)
	assert.Equal(t, ".total", ex.Endpoints[0].Metrics[1].Selector)

	assert.Equal(t, "https://reqres.in/api/apps", ex.Endpoints[1].Url)
	assert.Equal(t, 1, len(ex.Endpoints[1].Metrics))
	assert.Equal(t, "app_call_count", ex.Endpoints[1].Metrics[0].Name)
	assert.Equal(t, "Number of app calls", ex.Endpoints[1].Metrics[0].Description)
	assert.Equal(t, "counter", ex.Endpoints[1].Metrics[0].Type)
	assert.Equal(t, ".data.calls", ex.Endpoints[1].Metrics[0].Selector)
}
