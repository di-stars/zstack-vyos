package utils

import (
	"bytes"
	"text/template"
	"os/exec"
	"syscall"
	"fmt"
	"github.com/pkg/errors"
)

type Bash struct {
	Command string
	PipeFail bool
	Arguments map[string]string
}

func (b *Bash) build() error {
	Assert(b.Command != "", "Command cannot be emptry string")

	if (b.Arguments != nil) {
		tmpl, err := template.New("script").Parse(b.Command)
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, b.Arguments)
		if err != nil {
			return err
		}

		b.Command = buf.String()
	}

	if b.PipeFail {
		b.Command = fmt.Sprintf("set -o pipefail; %s", b.Command)
	}

	return nil
}

func (b *Bash) Run() error {
	if err := b.build(); err != nil {
		return err
	}

	ret, so, se, err := b.RunWithReturn()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to execute the command[%s] because of an internal errro",  b.Command))
	}

	if ret != 0 {
		return errors.New(fmt.Sprintf("failed to exectue the command[%s]\nreturn code:%d\nstdout:%s\nstderr:%s\n",
			b.Command, ret, so, se))
	}

	return nil
}

func (b *Bash) RunWithReturn() (retCode int, stdout, stderr string, err error) {
	if err = b.build(); err != nil {
		return -1, "", "", err
	}

	var so, se bytes.Buffer
	cmd := exec.Command("bash", "-c", b.Command)
	cmd.Stdout = &so
	cmd.Stderr = &se

	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			retCode = waitStatus.ExitStatus()
		} else {
			panic(errors.New("unable to get return code"))
		}
	} else {
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		retCode = waitStatus.ExitStatus()
	}

	stdout = string(so.Bytes())
	stderr = string(se.Bytes())

	return
}

func NewBash() *Bash {
	return &Bash{}
}

