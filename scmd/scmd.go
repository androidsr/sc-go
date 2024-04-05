package scmd

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/androidsr/sc-go/sc"
)

var (
	mutex sync.Mutex
	pwd   string
)

func init() {
	pwd = sc.GetExecuteDir()
}

type Command struct {
	sysType  string
	callback func(shell, output string) bool
	cmd      *exec.Cmd
	dir      string
}

func New(callback func(shell, output string) bool) *Command {
	c := &Command{}
	c.callback = callback
	c.sysType = runtime.GOOS
	return c
}

func (m *Command) Stop() {
	m.cmd.Process.Kill()
}

func IsDir(directoryPath string) bool {
	_, err := os.Stat(directoryPath)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func (m *Command) Command(shell string) error {
	defer func() {
		recover()
		os.Chdir(pwd)
	}()
	m.callback(shell, shell)
	if strings.HasPrefix(shell, "cd ") {
		m.dir = strings.TrimSpace(shell[3:])
		if !IsDir(m.dir) {
			return errors.New("目录不存在:" + m.dir)
		}
		return nil
	}
	mutex.Lock()
	defer mutex.Unlock()
	os.Chdir(m.dir)
	defer os.Chdir(pwd)
	if m.sysType == "linux" {
		newSh := strings.Fields(shell)
		m.cmd = exec.Command(newSh[1], newSh[1:]...)
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
		m.cmd = exec.Command(cmdName, args...)
	}
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		m.callback(shell, err.Error())
		return err
	}
	defer stdout.Close()
	stderr, _ := m.cmd.StderrPipe()
	err = m.cmd.Start()
	if err != nil {
		m.callback(shell, err.Error())
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
	res := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res = exitErr.ExitCode()
			stderr := string(exitErr.Stderr)
			m.callback(shell, stderr+exitErr.Error())
		} else {
			m.callback(shell, err.Error())
		}
	}

	if res == 0 {
		return nil
	} else {
		bs, _ := io.ReadAll(stderr)
		stderr.Close()
		m.callback(shell, err.Error())
		return errors.New(string(bs))
	}
}
