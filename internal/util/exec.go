package util

import (
	"context"
	"fmt"
	"github.com/markusressel/fan2go/internal/ui"
	"os/exec"
	"strings"
	"time"
)

func SafeCmdExecution(executable string, args []string, timeout time.Duration) (string, error) {
	if _, err := CheckFilePermissionsForExecution(executable); err != nil {
		return "", fmt.Errorf("cannot execute %s: %s", executable, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, executable, args...)
	out, err := cmd.Output()

	if ctx.Err() == context.DeadlineExceeded {
		ui.Warning("Command timed out: %s", executable)
		return "", err
	}

	if err != nil {
		exitError := err.(*exec.ExitError)
		ui.Warning("Command failed to execute: %s: %s", executable, string(exitError.Stderr))
		return "", err
	}

	strout := string(out)
	strout = strings.Trim(strout, "\n")

	return strout, nil
}
