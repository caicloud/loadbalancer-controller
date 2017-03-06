package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/golang/glog"
)

type keepalived struct {
	cmd *exec.Cmd
}

// Start starts a keepalived process in foreground.
// In case of any error it will terminate the execution with a fatal error
func (k *keepalived) Start() {
	k.cmd = exec.Command("keepalived",
		"--dont-fork",
		"--log-console",
		"--release-vips",
		"--pid", "/keepalived.pid")

	k.cmd.Stdout = os.Stdout
	k.cmd.Stderr = os.Stderr

	if err := k.cmd.Run(); err != nil {
		glog.Errorf("keepalived error: %v", err)
	}
}

// Reload sends SIGHUP to keepalived to reload the configuration.
func (k *keepalived) Reload() error {
	glog.Info("reloading keepalived")
	err := syscall.Kill(k.cmd.Process.Pid, syscall.SIGHUP)
	if err != nil {
		return fmt.Errorf("error reloading keepalived: %v", err)
	}

	return nil
}

// Stop stops keepalived process
func (k *keepalived) Stop() error {
	err := syscall.Kill(k.cmd.Process.Pid, syscall.SIGTERM)
	if err != nil {
		fmt.Errorf("error stopping keepalived: %v", err)
	}
	return nil
}
