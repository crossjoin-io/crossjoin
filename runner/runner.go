package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/crossjoin-io/crossjoin/api"
)

// Runner polls for tasks and executes them in
// containers.
type Runner struct {
	apiURL string
}

// NewRunner returns a new runner instance.
func NewRunner(apiURL string) (*Runner, error) {
	return &Runner{
		apiURL: apiURL,
	}, nil
}

func (run *Runner) Start() error {
	err := testDocker()
	if err != nil {
		return fmt.Errorf("test docker: %w -- Is Docker running?", err)
	}
	for {
		task, err := run.pollForTask()
		if err != nil {
			log.Fatal(err)
		}
		if task == nil {
			time.Sleep(5 * time.Second)
			continue
		}
		start := time.Now()
		result, err := run.runTaskContainer(task)
		if err != nil {
			return err
		}
		log.Println("dur", time.Since(start))

		body := &bytes.Buffer{}
		err = json.NewEncoder(body).Encode(result)
		if err != nil {
			return err
		}

		http.Post(run.apiURL+"/api/tasks/result", "application/json", body)
	}
}

func (run *Runner) pollForTask() (*api.Task, error) {
	resp, err := http.Get(run.apiURL + "/api/tasks/poll")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		// Didn't get a 2xx status code
		return nil, fmt.Errorf("got status %d", resp.StatusCode)
	}
	type apiResponse struct {
		OK       bool      `json:"ok"`
		Response *api.Task `json:"response"`
	}

	decodedResponse := apiResponse{}
	err = json.NewDecoder(resp.Body).Decode(&decodedResponse)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if !decodedResponse.OK {
		log.Println(decodedResponse)
		return nil, errors.New("api returned not ok")
	}
	return decodedResponse.Response, nil
}

func (run *Runner) downloadDataset(dataset string, destinationDirectory string) error {
	resp, err := http.Get(run.apiURL + fmt.Sprintf("/api/datasets/%s/download", dataset))
	if err != nil {
		log.Println(err)
		return err
	}
	if resp.StatusCode/100 != 2 {
		// Didn't get a 2xx status code
		return fmt.Errorf("got status %d", resp.StatusCode)
	}
	f, err := os.Create(filepath.Join(destinationDirectory, dataset+".db"))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (run *Runner) runTaskContainer(t *api.Task) (*api.TaskResult, error) {
	// Check the env params
	// TODO: make this safer
	for _, v := range t.Env {
		expanded := os.ExpandEnv(v)
		if len(expanded) == 0 {
			continue
		}
		// Quoted
		if (expanded[0] == '\'' && expanded[len(expanded)-1] == '\'') ||
			(expanded[0] == '"' && expanded[len(expanded)-1] == '"') {
			continue
		}
		if strings.ContainsAny(expanded, "\t ") {
			return nil, fmt.Errorf("env %s is not supported because it has spaces", v)
		}
	}

	dir, err := os.MkdirTemp("", "crossjoin_runner_*")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer os.RemoveAll(dir)
	input, err := json.Marshal(t.Input)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, dataset := range t.Datasets {
		// Download each dataset
		err = run.downloadDataset(dataset, dir)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	err = os.WriteFile(filepath.Join(dir, "in.json"), input, 0644)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	args := []string{"run", "--rm", "-v", fmt.Sprintf("%s:/runner/", dir)}
	if t.Script != "" {
		args = append(args, "--entrypoint", "/runner/entrypoint.sh")
		err = os.WriteFile(filepath.Join(dir, "entrypoint.sh"), []byte(t.Script), 0755)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	for k, v := range t.Env {
		v = os.ExpandEnv(v)
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, t.Image)
	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	resultOK := true
	if err != nil {
		log.Println(err)
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, err
		}
		resultOK = false
	}
	if stdout.Len() > 512 {
		stdout.Truncate(512)
	}
	if stderr.Len() > 512 {
		stderr.Truncate(512)
	}
	var output map[string]interface{}
	outputFile, err := os.Open(filepath.Join(dir, "out.json"))
	if err == nil {
		defer outputFile.Close()
		err = json.NewDecoder(outputFile).Decode(&output)
		if err != nil {
			return nil, err
		}
	}
	return &api.TaskResult{
		ID:     t.ID,
		OK:     resultOK,
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Output: output,
	}, nil
}

func testDocker() error {
	cmd := exec.Command("docker", "ps")
	_, err := cmd.CombinedOutput()
	return err
}
