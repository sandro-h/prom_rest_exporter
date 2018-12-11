package spec

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"vary/prom_rest_exporter/jq"
)

type ExporterSpec struct {
	Endpoints []*EndpointSpec
}

type EndpointSpec struct {
	Url     string
	Metrics []*MetricSpec
}

type MetricSpec struct {
	Name        string
	Description string
	Type        string
	Selector    string
	JqInst      *jq.Jq `yaml:"-"`
}

func (es EndpointSpec) String() string {
	data, _ := yaml.Marshal(es)
	return string(data)
}

func ReadSpecFromYamlFile(path string) (*ExporterSpec, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ex := new(ExporterSpec)
	err = yaml.Unmarshal(data, ex)
	if err != nil {
		return nil, err
	}

	ex.compileMetrics()
	if err != nil {
		return nil, err
	}

	return ex, nil
}

func (ex *ExporterSpec) compileMetrics() error {
	for _, e := range ex.Endpoints {
		for _, m := range e.Metrics {
			m.JqInst = jq.New()
			err := m.JqInst.CompileProgram(m.Selector)
			if err != nil {
				fmt.Printf("Metric compile error: %s\n", err.Error())
				// Handle somehow, either fail fast or ignore and expulse metric
			}
		}
	}

	return nil
}
