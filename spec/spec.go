package spec

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"vary/prom_rest_exporter/jq"
)

type EndpointSpec struct {
	Port    int
	Targets []*TargetSpec
}

type TargetSpec struct {
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

func (es TargetSpec) String() string {
	data, _ := yaml.Marshal(es)
	return string(data)
}

func ReadSpecFromYamlFile(path string) ([]*EndpointSpec, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var endpoints []*EndpointSpec
	err = yaml.Unmarshal(data, &endpoints)
	if err != nil {
		return nil, err
	}

	compileMetrics(endpoints)
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}

func compileMetrics(endpoints []*EndpointSpec) error {
	for _, e := range endpoints {
		for _, t := range e.Targets {
			for _, m := range t.Metrics {
				m.JqInst = jq.New()
				err := m.JqInst.CompileProgram(m.Selector)
				if err != nil {
					fmt.Printf("Metric compile error: %s\n", err.Error())
					// Handle somehow, either fail fast or ignore and expulse metric
				}
			}
		}
	}

	return nil
}
