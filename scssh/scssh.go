package scssh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Cli struct {
	IP         string      //IP地址
	Username   string      //用户名
	Password   string      //密码
	Port       int         //端口号
	client     *ssh.Client //ssh客户端
	LastResult string      //最近一次Run的结果
}

// 创建命令行对象
// @param ip IP地址
// @param username 用户名
// @param password 密码
// @param port 端口号,默认22
func New(ip string, username string, password string, port ...int) *Cli {
	cli := new(Cli)
	cli.IP = ip
	cli.Username = username
	cli.Password = password
	if len(port) <= 0 {
		cli.Port = 22
	} else {
		cli.Port = port[0]
	}
	return cli
}

func (c Cli) Close() error {
	return c.client.Close()
}

// 连接
func (c *Cli) connect() error {
	config := ssh.ClientConfig{
		User:            c.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(c.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以, 但是不够安全
		Timeout:         10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", c.IP, c.Port)
	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return err
	}
	c.client = sshClient
	return nil
}

// 执行shell
// @param shell shell脚本命令
func (c Cli) Run(shell string) (string, error) {
	if c.client == nil {
		if err := c.connect(); err != nil {
			return "", err
		}
	}
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	buf, err := session.CombinedOutput(shell)

	c.LastResult = string(buf)
	return c.LastResult, err
}

// ssh 远程命令执行
func (c Cli) NewTerminal() (*Terminal, error) {
	if c.client == nil {
		if err := c.connect(); err != nil {
			return nil, err
		}
	}
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	in, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = session.Shell(); err != nil {
		return nil, err
	}
	terminal := &Terminal{cli: c, session: session, input: in, output: bufio.NewReader(out)}
	return terminal, nil
}

type Terminal struct {
	cli     Cli
	pid     string
	session *ssh.Session
	input   io.WriteCloser
	output  *bufio.Reader
}

func (t Terminal) Write(shell string) {
	t.input.Write([]byte(shell + ";echo $?;echo sc\n"))
}

func (t Terminal) Read(wr io.WriteCloser) error {
	var state string
	for {
		line, err := t.output.ReadString('\n')
		if err != nil || strings.TrimSpace(line) == "sc" {
			break
		}
		state = strings.TrimSpace(state)
		wr.Write([]byte(strings.TrimSpace(line) + "\n"))
	}
	if state != "0" {
		return errors.New("命令执行失败")
	}
	return nil
}

func (t Terminal) Close() error {
	_, err := t.cli.Run("kill -9 -" + t.pid)
	return err
}
