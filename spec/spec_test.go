package spec

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadSpecFromYaml(t *testing.T) {
	endpoints, err := ReadSpecFromYamlFile("spec_test_spec.yml")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(endpoints))

	endpoint := endpoints[0]
	assert.Equal(t, 9011, endpoint.Port)
	assert.Equal(t, "https://reqres.in/api/users", endpoint.Targets[0].Url)
	assert.Equal(t, 2, len(endpoint.Targets[0].Metrics))
	assert.Equal(t, "user_count", endpoint.Targets[0].Metrics[0].Name)
	assert.Equal(t, "Number of users", endpoint.Targets[0].Metrics[0].Description)
	assert.Equal(t, "gauge", endpoint.Targets[0].Metrics[0].Type)
	assert.Equal(t, "[.data[].last_name] | length", endpoint.Targets[0].Metrics[0].Selector)
	assert.NotNil(t, endpoint.Targets[0].Metrics[0].JqInst)

	assert.Equal(t, "user_count_total", endpoint.Targets[0].Metrics[1].Name)
	assert.Equal(t, "Total number of users", endpoint.Targets[0].Metrics[1].Description)
	assert.Equal(t, "gauge", endpoint.Targets[0].Metrics[1].Type)
	assert.Equal(t, ".total", endpoint.Targets[0].Metrics[1].Selector)
	assert.NotNil(t, endpoint.Targets[0].Metrics[1].JqInst)

	assert.Equal(t, "https://reqres.in/api/apps", endpoint.Targets[1].Url)
	assert.Equal(t, 1, len(endpoint.Targets[1].Metrics))
	assert.Equal(t, "app_call_count", endpoint.Targets[1].Metrics[0].Name)
	assert.Equal(t, "Number of app calls", endpoint.Targets[1].Metrics[0].Description)
	assert.Equal(t, "counter", endpoint.Targets[1].Metrics[0].Type)
	assert.Equal(t, ".data.calls", endpoint.Targets[1].Metrics[0].Selector)
	assert.NotNil(t, endpoint.Targets[1].Metrics[0].JqInst)
}
