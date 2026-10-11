//go:build unix

package mobile

import (
	"errors"
	"os/exec"
	"syscall"
)

func stoppedBySignal(err error) bool {
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		return false
	}
	ws, ok := ee.Sys().(syscall.WaitStatus)
	return ok && ws.Signaled()
}
