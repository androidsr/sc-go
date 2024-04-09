package scmd

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/androidsr/sc-go/sc"
)

var (
	pwd string
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
	if m.sysType == "linux" {
		if (!strings.HasPrefix(shell, "sh ") && (strings.HasSuffix(strings.TrimSpace(shell), ".sh") ||
			strings.Contains(strings.TrimSpace(shell), ".sh "))) || strings.Contains(shell, " && ") ||
			strings.HasPrefix(strings.TrimSpace(shell), "mv ") || strings.HasPrefix(strings.TrimSpace(shell), "cp ") ||
			strings.HasPrefix(strings.TrimSpace(shell), "rm ") || strings.HasPrefix(strings.TrimSpace(shell), "tar ") ||
			strings.HasPrefix(strings.TrimSpace(shell), "unzip ") {
			m.cmd = exec.Command("bash", "-c", shell)
		} else {
			newSh := strings.Fields(shell)
			m.cmd = exec.Command(newSh[0], newSh[1:]...)
		}
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
	m.cmd.Dir = m.dir

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
	defer stderr.Close()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m.callback(shell, scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			m.callback(shell, scanner.Text())
		}
	}()
	if err := m.cmd.Wait(); err != nil {
		m.callback(shell, err.Error())
	}
	return nil
}
