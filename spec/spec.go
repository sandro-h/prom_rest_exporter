package spec

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"vary/prom_rest_exporter/jq"
)

type ExporterSpec struct {
	Endpoints        []*EndpointSpec
	CacheTimeSeconds int `yaml:"cache_time"`
}

type EndpointSpec struct {
	Port             int
	Targets          []*TargetSpec
	CacheTimeSeconds int `yaml:"cache_time"`
}

type TargetSpec struct {
	URL     string
	Metrics []*MetricSpec
}

type MetricSpec struct {
	Name        string
	Description string
	Type        string
	Selector    string
	ValSelector string `yaml:"val_selector"`
	Labels      []*LabelSpec
	// Calculated fields:
	OnlyFixedLabels bool   `yaml:"-"`
	JqInst          *jq.Jq `yaml:"-"`
	ValJqInst       *jq.Jq `yaml:"-"`
}

type LabelSpec struct {
	Name       string
	Selector   string
	FixedValue string `yaml:"fixed_value"`
	JqInst     *jq.Jq `yaml:"-"`
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

	return readSpec(&data)
}

func ReadSpecFromYamlString(yaml string) (*ExporterSpec, error) {
	bytes := []byte(yaml)
	return readSpec(&bytes)
}

func readSpec(yamlBytes *[]byte) (*ExporterSpec, error) {
	var ex ExporterSpec
	err := yaml.Unmarshal(*yamlBytes, &ex)
	if err != nil {
		return nil, err
	}

	err = validateSpec(&ex)
	if err != nil {
		return nil, err
	}

	err = compileJqsInSpec(&ex)
	if err != nil {
		return nil, err
	}

	return &ex, nil
}

func validateSpec(ex *ExporterSpec) error {
	return ex.Validate()
}

func compileJqsInSpec(ex *ExporterSpec) error {
	var err error
	for _, e := range ex.Endpoints {
		for _, t := range e.Targets {
			for _, m := range t.Metrics {
				// Compile metric selectors
				m.JqInst, err = compileJq(m.Selector)
				if err != nil {
					return err
				}

				// Compile metric value selectors
				if m.ValSelector != "" && m.ValSelector != "." {
					m.ValJqInst, err = compileJq(m.ValSelector)
					if err != nil {
						return err
					}
				}

				// Compile metric label selectors
				m.OnlyFixedLabels = true
				for _, l := range m.Labels {
					if l.FixedValue == "" && l.Selector != "" {
						l.JqInst, err = compileJq(l.Selector)
						if err != nil {
							return err
						}
						m.OnlyFixedLabels = false
					}
				}
			}
		}
	}

	return nil
}

func compileJq(selector string) (*jq.Jq, error) {
	jqInst := jq.New()
	err := jqInst.CompileProgram(selector)
	if err != nil {
		msg := fmt.Sprintf("Jq compile error for selector %s: %s\n", selector, err.Error())
		return nil, errors.New(msg)
	}
	return jqInst, nil
}

func (s *ExporterSpec) Validate() error {
	for _, ep := range s.Endpoints {
		err := ep.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *EndpointSpec) Validate() error {
	if s.Port <= 0 {
		return errors.New("Endpoint 'port' must be > 0")
	}

	for _, t := range s.Targets {
		err := t.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TargetSpec) Validate() error {
	if s.URL == "" {
		return errors.New("Target must have 'url'")
	}
	for _, m := range s.Metrics {
		err := m.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MetricSpec) Validate() error {
	if s.Name == "" {
		return errors.New("Metric must have 'name'")
	}
	if s.Selector == "" {
		return errors.New("Metric must have 'selector'")
	}
	for _, l := range s.Labels {
		err := l.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *LabelSpec) Validate() error {
	if s.Name == "" {
		return errors.New("Label must have 'name'")
	}
	if s.Selector == "" && s.FixedValue == "" {
		return errors.New("Label must have 'selector' or 'fixed_value'")
	}
	return nil
}
