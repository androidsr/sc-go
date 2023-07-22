package scmd

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Command struct {
	sysType  string
	callback func(shell, output string) bool
	cmd      *exec.Cmd
	dir      string
	isRun    bool
	waitRun  bool
}

func New(callback func(shell, output string) bool, stopBack func(output string)) *Command {
	c := &Command{}
	c.waitRun = true
	c.isRun = true
	c.WaitRun(stopBack)
	c.callback = callback
	c.sysType = runtime.GOOS
	return c
}

func (m *Command) WaitRun(callback func(output string)) {
	go func() {
		for m.waitRun {
			time.Sleep(time.Second * time.Duration(2))
			if !m.isRun && m.waitRun {
				m.cmd.Process.Kill()
				callback("======\n终止命令")
				break
			}
		}
	}()
}

func (m *Command) Command(shell string) error {
	defer func() {
		recover()
	}()
	if strings.HasPrefix(shell, "cd ") {
		m.dir = strings.TrimSpace(shell[3:])
		return nil
	}
	if m.sysType == "linux" {
		newSh := shell
		if m.dir != "" {
			newSh = "cd " + m.dir + " && " + shell
		}
		m.cmd = exec.Command("bash", "-c", newSh)
	} else if m.sysType == "windows" {
		var cmdName string
		args := make([]string, 0)
		for i, v := range strings.Split(shell, " ") {
			if i == 0 {
				cmdName = v
			} else if v != "" && v != " " {
				args = append(args, strings.TrimSpace(v))
			}
		}
		if m.dir != "" {
			cmdName = "cd " + m.dir + " && " + cmdName
		}
		m.cmd = exec.Command(cmdName, args...)
	}
	stderr := bytes.Buffer{}
	m.cmd.Stderr = &stderr
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()
	err = m.cmd.Start()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(stdout)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					m.callback(shell, err.Error())
				}
				break
			}
			result := m.callback(shell, line)
			if !result {
				break
			}
		}
	}()
	err = m.cmd.Wait()
	m.waitRun = false
	m.isRun = false
	if err != nil {
		return err
	}
	if stderr.String() == "" {
		return nil
	} else {
		return errors.New(stderr.String())
	}
}
