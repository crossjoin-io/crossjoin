package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DataSets  []DataSet `yaml:"data_sets"`
	Workflows []Workflow
}

type DataSet struct {
	Name       string      `yaml:"name"`
	DataSource *DataSource `yaml:"data_source"`
	Joins      []Join      `yaml:"joins"`
}

type DataSource struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	Path             string `yaml:"path"`
	ConnectionString string `yaml:"connection_string"`
	Query            string `yaml:"query"`
}

func (ds *DataSource) expandConnectionString() {
	if strings.HasPrefix(ds.ConnectionString, "$") {
		ds.ConnectionString = os.ExpandEnv(ds.ConnectionString)
	}
}

type Join struct {
	Type       string        `yaml:"type"`
	Columns    []JoinColumns `yaml:"columns"`
	DataSource *DataSource   `yaml:"data_source"`
}

type JoinColumns struct {
	LeftColumn  string `yaml:"left_column"`
	RightColumn string `yaml:"right_column"`
}

type Workflow struct {
	ID    string                   `yaml:"id"`
	Start string                   `yaml:"start"`
	Tasks map[string]*WorkflowTask `yaml:"tasks"`
}

func (w *Workflow) Parse(content []byte) error {
	return yaml.Unmarshal(content, w)
}

type WorkflowTask struct {
	Next string `yaml:"next,omitempty"`

	Type   string                 `yaml:"type"`
	Params map[string]interface{} `yaml:"params"`

	Image  string `yaml:"image,omitempty"` // for "container" type
	Script string `yaml:"script,omitempty"`
}

func (c *Config) Parse(content []byte) error {
	err := yaml.Unmarshal(content, c)
	if err != nil {
		return err
	}

	for _, dataset := range c.DataSets {
		if dataset.DataSource != nil {
			dataset.DataSource.expandConnectionString()
		}
		for _, j := range dataset.Joins {
			if j.DataSource != nil {
				j.DataSource.expandConnectionString()
			}
		}
	}

	return c.validate()
}

func (c *Config) String() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

func (w Workflow) String() string {
	b, _ := yaml.Marshal(w)
	return string(b)
}

func (c *Config) validate() error {
	seenDataSetNames := map[string]bool{}
	for _, dataSet := range c.DataSets {
		seenDataSourceNames := map[string]bool{}
		if !validName(dataSet.Name) {
			return fmt.Errorf("invalid name `%s`", dataSet.Name)
		}
		if seenDataSetNames[dataSet.Name] {
			return fmt.Errorf("duplicate data set name `%s`", dataSet.Name)
		}
		seenDataSetNames[dataSet.Name] = true
		if dataSet.DataSource == nil {
			return errors.New("missing data source")
		}
		err := dataSet.DataSource.validate()
		if err != nil {
			return err
		}
		if dataSet.Name == dataSet.DataSource.Name {
			return fmt.Errorf("data source can't have the same name as the data set (`%s`)", dataSet.Name)
		}
		seenDataSourceNames[dataSet.DataSource.Name] = true
		for _, j := range dataSet.Joins {
			if j.DataSource == nil {
				return errors.New("missing data source for join")
			}
			err := j.DataSource.validate()
			if err != nil {
				return err
			}
			if seenDataSourceNames[j.DataSource.Name] {
				return fmt.Errorf("duplicate data source name `%s`", j.DataSource.Name)
			}
			seenDataSetNames[j.DataSource.Name] = true
		}
	}
	return nil
}

func (ds *DataSource) validate() error {
	if !validName(ds.Name) {
		return fmt.Errorf("invalid name `%s`", ds.Name)
	}
	switch ds.Type {
	case "postgres":
		if ds.Query == "" {
			return fmt.Errorf("missing query for data source `%s`", ds.Name)
		}
	case "csv":
		if ds.Path == "" {
			return fmt.Errorf("missing path for data source `%s`", ds.Name)
		}
	default:
		return fmt.Errorf("unknown data source type `%s`", ds.Type)
	}
	return nil
}

var validNameRegexp = regexp.MustCompile(`^[a-zA-Z]([\w-]*[a-zA-Z0-9])?$`)

func validName(s string) bool {
	return validNameRegexp.MatchString(s)
}
