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
	Port    int
	Targets []*TargetSpec
}

type TargetSpec struct {
	Url     string
	Metrics []*MetricSpec
}

type MetricSpec struct {
	Name         string
	Description  string
	Type         string
	Selector     string
	Val_Selector string
	Labels       []*LabelSpec
	JqInst       *jq.Jq `yaml:"-"`
	ValJqInst    *jq.Jq `yaml:"-"`
}

type LabelSpec struct {
	Name     string
	Selector string
	JqInst   *jq.Jq `yaml:"-"`
}

func (es TargetSpec) String() string {
	data, _ := yaml.Marshal(es)
	return string(data)
}

func ReadSpecFromYamlFile(path string) (*ExporterSpec, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ex ExporterSpec
	err = yaml.Unmarshal(data, &ex)
	if err != nil {
		return nil, err
	}

	err = validateSpec(&ex)
	if err != nil {
		return nil, err
	}

	compileJqsInSpec(&ex)
	if err != nil {
		return nil, err
	}

	return &ex, nil
}

func validateSpec(ex *ExporterSpec) error {
	// TODO
	return nil
}

func compileJqsInSpec(ex *ExporterSpec) error {
	for _, e := range ex.Endpoints {
		for _, t := range e.Targets {
			for _, m := range t.Metrics {
				m.JqInst = compileJq(m.Selector)
				if m.Val_Selector != "" && m.Val_Selector != "." {
					m.ValJqInst = compileJq(m.Val_Selector)
				}
				for _, l := range m.Labels {
					l.JqInst = compileJq(l.Selector)
				}
			}
		}
	}

	return nil
}

func compileJq(selector string) *jq.Jq {
	jqInst := jq.New()
	err := jqInst.CompileProgram(selector)
	if err != nil {
		fmt.Printf("Jq compile error: %s\n", err.Error())
		// Handle somehow, either fail fast or ignore and expulse metric
	}
	return jqInst
}
