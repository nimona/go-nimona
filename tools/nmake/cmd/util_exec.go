package cmd

import (
	"os"
	osexec "os/exec"
)

// exec a command and return its output and error
func exec(env []string, cmd string, args []string) (string, error) {
	res := osexec.Command(cmd, args...)
	res.Env = append(os.Environ(), env...)
	b, err := res.Output()
	return string(b), err
}

// execPipe a command and pipes the stdout and stderr outputs to the given files
func execPipe(env []string, cmd string, args []string, stdout, stderr *os.File) error {
	res := osexec.Command(cmd, args...)
	res.Env = append(os.Environ(), env...)
	res.Stdout = stdout
	res.Stderr = stderr
	return res.Run()
}
