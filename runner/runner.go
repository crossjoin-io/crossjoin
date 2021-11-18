package runner

// Runner polls for tasks and executes them in
// containers.
type Runner struct {
	pollAddress string
}

// NewRunner returns a new runner instance.
func NewRunner(pollAddress string) (*Runner, error) {
	return &Runner{
		pollAddress: pollAddress,
	}, nil
}

func (r *Runner) Start() error {
	return nil
}
