package spec

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadSpecFromYaml(t *testing.T) {
	spec, err := ReadSpecFromYamlFile("testdata/spec_test_spec.yml")
	assert.Nil(t, err)
	assert.Equal(t, 60, spec.CacheTimeSeconds)
	assert.Equal(t, 1, len(spec.Endpoints))

	endpoint := spec.Endpoints[0]
	assert.Equal(t, 9011, endpoint.Port)
	assert.Equal(t, 30, endpoint.CacheTimeSeconds)
	assert.Equal(t, "https://reqres.in/api/users", endpoint.Targets[0].URL)
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
	assert.Equal(t, 1, len(endpoint.Targets[0].Metrics[1].Labels))
	assert.Equal(t, "instance", endpoint.Targets[0].Metrics[1].Labels[0].Name)
	assert.Equal(t, ".inst", endpoint.Targets[0].Metrics[1].Labels[0].Selector)
	assert.NotNil(t, endpoint.Targets[0].Metrics[1].Labels[0].JqInst)

	assert.Equal(t, "https://reqres.in/api/apps", endpoint.Targets[1].URL)
	assert.Equal(t, 1, len(endpoint.Targets[1].Metrics))
	assert.Equal(t, "app_call_count", endpoint.Targets[1].Metrics[0].Name)
	assert.Equal(t, "Number of app calls", endpoint.Targets[1].Metrics[0].Description)
	assert.Equal(t, "counter", endpoint.Targets[1].Metrics[0].Type)
	assert.Equal(t, ".data.calls", endpoint.Targets[1].Metrics[0].Selector)
	assert.NotNil(t, endpoint.Targets[1].Metrics[0].JqInst)
}

func TestReadSpecFromInexistentFile(t *testing.T) {
	spec, err := ReadSpecFromYamlFile("testdata/does_not_exist.yml")
	assert.Nil(t, spec)
	assert.NotNil(t, err)
}

func TestReadSpecWithInvalidYaml(t *testing.T) {
	spec, err := ReadSpecFromYamlFile("testdata/spec_test_invalid_yaml_spec.yml")
	assert.Nil(t, spec)
	assert.NotNil(t, err)
	assert.Equal(t, "yaml: line 4: did not find expected '-' indicator", err.Error())
}

func TestReadSpecWithInvalidStructure(t *testing.T) {
	spec, err := ReadSpecFromYamlFile("testdata/spec_test_invalid_struct_spec.yml")
	assert.Nil(t, spec)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "line 8: cannot unmarshal !!map into []*spec.MetricSpec")
}

func TestReadSpecWithInvalidJq(t *testing.T) {
	spec, err := ReadSpecFromYamlFile("testdata/spec_test_invalid_jq_spec.yml")
	assert.Nil(t, spec)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Jq compile error for selector [.data[].last_name | length: jq: error: syntax error, unexpected $end")
}
