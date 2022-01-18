package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Datasets        []Dataset        `yaml:"datasets"`
	DataConnections []DataConnection `yaml:"data_connections"`
	Workflows       []Workflow       `yaml:"workflows"`
}

type Dataset struct {
	Name       string      `yaml:"name"`
	Refresh    *Refresh    `yaml:"refresh"`
	DataSource *DataSource `yaml:"data_source"`
	Joins      []Join      `yaml:"joins"`
}

type Refresh struct {
	Interval string `yaml:"interval"`
}

type DataConnection struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	Path             string `yaml:"path"`
	ConnectionString string `yaml:"connection_string"`
}

type DataSource struct {
	Name           string `yaml:"name"`
	DataConnection string `yaml:"data_connection"`
	Query          string `yaml:"query"`
}

func (dc *DataConnection) expandConnectionString() {
	if strings.HasPrefix(dc.ConnectionString, "$") {
		dc.ConnectionString = os.ExpandEnv(dc.ConnectionString)
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
	On    *WorkflowTrigger         `yaml:"on"`
	Tasks map[string]*WorkflowTask `yaml:"tasks"`
}

type WorkflowTrigger struct {
	DatasetRefresh []string `yaml:"dataset_refresh"`
}

func (w *Workflow) Parse(content []byte) error {
	return yaml.Unmarshal(content, w)
}

type WorkflowTask struct {
	Next string `yaml:"next,omitempty"`

	Type         string                 `yaml:"type"`
	Env          map[string]string      `yaml:"env"`
	With         map[string]interface{} `yaml:"with"`
	WithDatasets []string               `yaml:"with_datasets"`

	Image  string `yaml:"image,omitempty"` // for "container" type
	Script string `yaml:"script,omitempty"`
}

func (c *Config) Parse(content []byte, dir string) error {
	err := yaml.Unmarshal(content, c)
	if err != nil {
		return err
	}

	for i, dataConnection := range c.DataConnections {
		dataConnection.expandConnectionString()
		if dataConnection.Path != "" {
			if !filepath.IsAbs(dataConnection.Path) {
				c.DataConnections[i].Path = filepath.Join(dir, dataConnection.Path)
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

	dataConnectionTypes := map[string]string{}
	seenDataConnectionNames := map[string]bool{}
	for _, dataConnection := range c.DataConnections {
		err := dataConnection.validate()
		if err != nil {
			return err
		}
		if seenDataConnectionNames[dataConnection.Name] {
			return fmt.Errorf("duplicate data connection name `%s`", dataConnection.Name)
		}
		seenDataConnectionNames[dataConnection.Name] = true
		dataConnectionTypes[dataConnection.Name] = dataConnection.Type
	}

	seenDataSetNames := map[string]bool{}
	for _, dataset := range c.Datasets {
		seenDataSourceNames := map[string]bool{}
		if !validName(dataset.Name) {
			return fmt.Errorf("invalid name `%s`", dataset.Name)
		}
		if seenDataSetNames[dataset.Name] {
			return fmt.Errorf("duplicate dataset name `%s`", dataset.Name)
		}
		seenDataSetNames[dataset.Name] = true
		if dataset.DataSource == nil {
			return errors.New("missing data source")
		}
		err := dataset.DataSource.validate(dataConnectionTypes[dataset.DataSource.DataConnection])
		if err != nil {
			return err
		}
		if dataset.Name == dataset.DataSource.Name {
			return fmt.Errorf("data source can't have the same name as the dataset (`%s`)", dataset.Name)
		}
		seenDataSourceNames[dataset.DataSource.Name] = true
		for _, j := range dataset.Joins {
			if j.DataSource == nil {
				return errors.New("missing data source for join")
			}
			err := j.DataSource.validate(dataConnectionTypes[j.DataSource.DataConnection])
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

func (dc *DataConnection) validate() error {
	if !validName(dc.Name) {
		return fmt.Errorf("invalid name `%s`", dc.Name)
	}
	switch dc.Type {
	case "postgres":
		if dc.ConnectionString == "" {
			return fmt.Errorf("missing connection string for data connection `%s`", dc.Name)
		}
	case "csv":
		if dc.Path == "" {
			return fmt.Errorf("missing path for data connection `%s`", dc.Name)
		}
	default:
		return fmt.Errorf("unknown data connection type `%s`", dc.Type)
	}
	return nil
}

func (ds *DataSource) validate(dataConnectionType string) error {
	if !validName(ds.Name) {
		return fmt.Errorf("invalid name `%s`", ds.Name)
	}
	switch dataConnectionType {
	case "postgres":
		if ds.Query == "" {
			return fmt.Errorf("missing query")
		}
	}
	return nil
}

var validNameRegexp = regexp.MustCompile(`^[a-zA-Z]([\w-]*[a-zA-Z0-9])?$`)

func validName(s string) bool {
	return validNameRegexp.MatchString(s)
}
