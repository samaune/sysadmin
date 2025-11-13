package exec

import (
	"bytes"
	"os/exec"
)

func RunCmd(command string) ([]byte, error) {
	return ExecCmd("cmd", "/C", command)
}

func ExecCmd(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.Bytes(), err
}
