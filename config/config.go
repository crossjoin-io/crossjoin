package config

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Datasets        []Dataset        `yaml:"datasets" json:"datasets"`
	DataConnections []DataConnection `yaml:"data_connections" json:"data_connections"`
	Workflows       []Workflow       `yaml:"workflows" json:"workflows"`
}

type Dataset struct {
	ID         string      `yaml:"id" json:"id"`
	Refresh    *Refresh    `yaml:"refresh" json:"refresh"`
	DataSource *DataSource `yaml:"data_source" json:"data_source"`
	Joins      []Join      `yaml:"joins" json:"joins"`
}

type Refresh struct {
	Interval string `yaml:"interval" json:"interval"`
}

type DataConnection struct {
	ID               string `yaml:"id" json:"id"`
	Type             string `yaml:"type" json:"type"`
	Path             string `yaml:"path" json:"path"`
	ConnectionString string `yaml:"connection_string" json:"connection_string"`
}

type DataSource struct {
	ID             string `yaml:"id" json:"id"`
	DataConnection string `yaml:"data_connection" json:"data_connection"`
	Query          string `yaml:"query" json:"query"`
}

func (dc *DataConnection) ExpandConnectionString() {
	if strings.HasPrefix(dc.ConnectionString, "$") {
		dc.ConnectionString = os.ExpandEnv(dc.ConnectionString)
	}
}

type Join struct {
	Type       string        `yaml:"type" json:"type"`
	Columns    []JoinColumns `yaml:"columns" json:"columns"`
	DataSource *DataSource   `yaml:"data_source" json:"data_source"`
}

type JoinColumns struct {
	LeftColumn  string `yaml:"left_column" json:"left_column"`
	RightColumn string `yaml:"right_column" json:"right_column"`
}

type Workflow struct {
	ID    string                   `yaml:"id" json:"id"`
	Start string                   `yaml:"start" json:"start"`
	On    *WorkflowTrigger         `yaml:"on" json:"on"`
	Tasks map[string]*WorkflowTask `yaml:"tasks" json:"tasks"`
}

type WorkflowTrigger struct {
	DatasetRefresh []string `yaml:"dataset_refresh" json:"dataset_refresh"`
}

func (w *Workflow) Parse(content []byte) error {
	return yaml.Unmarshal(content, w)
}

type WorkflowTask struct {
	Next string `yaml:"next,omitempty" json:"next,omitempty"`

	Type         string                 `yaml:"type" json:"type"`
	Env          map[string]string      `yaml:"env" json:"env"`
	With         map[string]interface{} `yaml:"with" json:"with"`
	WithDatasets []string               `yaml:"with_datasets" json:"with_datasets"`

	Image  string `yaml:"image,omitempty" json:"image,omitempty"` // for "container" type
	Script string `yaml:"script,omitempty" json:"script,omitempty"`
}

func (c *Config) Parse(content []byte, dir string) error {
	err := yaml.Unmarshal(content, c)
	if err != nil {
		return err
	}

	urlDir, _ := url.Parse(dir)

	for i, dataConnection := range c.DataConnections {
		dataConnection.ExpandConnectionString()
		if dataConnection.Path != "" {
			if !path.IsAbs(dataConnection.Path) {
				if urlDir != nil {
					urlPath := *urlDir
					urlPath.Path = path.Join(urlPath.Path, dataConnection.Path)
					c.DataConnections[i].Path = urlPath.String()
				} else {
					c.DataConnections[i].Path = filepath.Join(dir, dataConnection.Path)
				}
			}
		}
	}

	return c.validate()
}

func (c *Config) String() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

func (c *Config) JSON() []byte {
	b, _ := json.Marshal(c)
	return b
}

func (c *Config) Hash() string {
	hash := sha1.Sum(c.JSON())
	return fmt.Sprintf("%x", hash)
}

func (w Workflow) String() string {
	b, _ := yaml.Marshal(w)
	return string(b)
}

func (c *Config) validate() error {

	dataConnectionTypes := map[string]string{}
	seenDataConnectionIDs := map[string]bool{}
	for _, dataConnection := range c.DataConnections {
		err := dataConnection.validate()
		if err != nil {
			return err
		}
		if seenDataConnectionIDs[dataConnection.ID] {
			return fmt.Errorf("duplicate data connection ID `%s`", dataConnection.ID)
		}
		seenDataConnectionIDs[dataConnection.ID] = true
		dataConnectionTypes[dataConnection.ID] = dataConnection.Type
	}

	seenDataSetIDs := map[string]bool{}
	for _, dataset := range c.Datasets {
		seenDataSourceIDs := map[string]bool{}
		if !validID(dataset.ID) {
			return fmt.Errorf("invalid ID `%s`", dataset.ID)
		}
		if seenDataSetIDs[dataset.ID] {
			return fmt.Errorf("duplicate dataset ID `%s`", dataset.ID)
		}
		seenDataSetIDs[dataset.ID] = true
		if dataset.DataSource == nil {
			return errors.New("missing data source")
		}
		err := dataset.DataSource.validate(dataConnectionTypes[dataset.DataSource.DataConnection])
		if err != nil {
			return err
		}
		if dataset.ID == dataset.DataSource.ID {
			return fmt.Errorf("data source can't have the same ID as the dataset (`%s`)", dataset.ID)
		}
		seenDataSourceIDs[dataset.DataSource.ID] = true
		for _, j := range dataset.Joins {
			if j.DataSource == nil {
				return errors.New("missing data source for join")
			}
			err := j.DataSource.validate(dataConnectionTypes[j.DataSource.DataConnection])
			if err != nil {
				return err
			}
			if seenDataSourceIDs[j.DataSource.ID] {
				return fmt.Errorf("duplicate data source ID `%s`", j.DataSource.ID)
			}
			seenDataSetIDs[j.DataSource.ID] = true
		}
	}
	return nil
}

func (dc *DataConnection) validate() error {
	if !validID(dc.ID) {
		return fmt.Errorf("invalid ID `%s`", dc.ID)
	}
	switch dc.Type {
	case "postgres":
		if dc.ConnectionString == "" {
			return fmt.Errorf("missing connection string for data connection `%s`", dc.ID)
		}
	case "csv":
		if dc.Path == "" {
			return fmt.Errorf("missing path for data connection `%s`", dc.ID)
		}
	default:
		return fmt.Errorf("unknown data connection type `%s`", dc.Type)
	}
	return nil
}

func (ds *DataSource) validate(dataConnectionType string) error {
	if !validID(ds.ID) {
		return fmt.Errorf("invalid ID `%s`", ds.ID)
	}
	switch dataConnectionType {
	case "postgres":
		if ds.Query == "" {
			return fmt.Errorf("missing query")
		}
	}
	return nil
}

var validIDRegexp = regexp.MustCompile(`^[a-zA-Z]([\w-]*[a-zA-Z0-9])?$`)

func validID(s string) bool {
	return validIDRegexp.MatchString(s)
}
