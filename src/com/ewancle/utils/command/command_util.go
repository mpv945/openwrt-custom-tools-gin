package command_util

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

type CmdResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

func ExecCommand(name string, args ...string) CmdResult {

	cmd := exec.Command(name, args...)

	// 设置超时
	/*ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ping", "google.com")*/

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var code int = 0
	if err != nil {
		if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
			code := exitError.ExitCode()
			fmt.Println("exit code:", code)
		}
	}

	return CmdResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: code,
		Error:    err,
	}
}
