package scmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

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
	isRun    bool
	waitRun  bool
}

func New(callback func(shell, output string) bool) *Command {
	c := &Command{}
	c.waitRun = true
	c.isRun = true
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
	os.Chdir(m.dir)
	if m.sysType == "linux" {
		newSh := shell
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
		m.cmd = exec.Command(cmdName, args...)
	}
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		os.Chdir(pwd)
		mutex.Unlock()
		return err
	}
	defer stdout.Close()
	stderr, _ := m.cmd.StderrPipe()
	err = m.cmd.Start()
	if err != nil {
		os.Chdir(pwd)
		mutex.Unlock()
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
	os.Chdir(pwd)
	mutex.Unlock()
	res := 0
	if err := m.cmd.Wait(); err != nil {
		fmt.Println(err)
		if ex, ok := err.(*exec.ExitError); ok {
			res = ex.Sys().(syscall.WaitStatus).ExitStatus() //获取命令执行返回状态，相当于shell: echo $?
		}
	}
	m.waitRun = false
	m.isRun = false
	if res == 0 {
		return nil
	} else {
		bs, _ := io.ReadAll(stderr)
		stderr.Close()
		return errors.New(string(bs))
	}
}
