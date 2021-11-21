package api

type Task struct {
	ID     string      `json:"id"`
	Image  string      `json:"image"`
	Script string      `json:"script"`
	Input  interface{} `json:"input"`
}

type TaskResult struct {
	ID     string      `json:"id"`
	OK     bool        `json:"ok"`
	Output interface{} `json:"output"`
	Stdout string      `json:"stdout"`
	Stderr string      `json:"stderr"`
}

type Response struct {
	OK       bool        `json:"ok"`
	Status   int         `json:"-"`
	Error    string      `json:"error,omitempty"`
	Response interface{} `json:"response,omitempty"`
}
