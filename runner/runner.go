package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	errors := make(chan error)
	for i := 0; i < 4; i++ {
		go func() {
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
				result, err := runTaskContainer(task)
				if err != nil {
					errors <- err
				}
				log.Println("got result", result)
				log.Println("dur", time.Since(start))

				body := &bytes.Buffer{}
				json.NewEncoder(body).Encode(result)

				http.Post(run.apiURL+"/api/tasks/result", "application/json", body)
			}
		}()
	}
	err := <-errors
	return err
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

func runTaskContainer(t *api.Task) (*api.TaskResult, error) {
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
	args = append(args, t.Image)
	fmt.Println(args)
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
	var output interface{}
	outputFile, err := os.Open(filepath.Join(dir, "out.json"))
	if err == nil {
		defer outputFile.Close()
		json.NewDecoder(outputFile).Decode(&output)
	}
	return &api.TaskResult{
		ID:     t.ID,
		OK:     resultOK,
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Output: output,
	}, nil
}
