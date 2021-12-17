package config

import "testing"

func TestParseWithWorkflow(t *testing.T) {
	conf := &Config{}
	err := conf.Parse([]byte(`
workflows:
  - id: my-workflow
    start: send_slack_message
    tasks:
      send_slack_message:
        type: container
        next: wait_for_1_day
        image: alpine
        params:
          foo: bar
        script: |
          #!/bin/sh
          apk add jq
          echo hi
      wait_for_1_day:
        type: delay
        params:
          duration: 1d`), "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(conf)
}
