package scssh

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Cli struct {
	IP         string      //IP地址
	Username   string      //用户名
	Password   string      //密码
	Port       int         //端口号
	client     *ssh.Client //ssh客户端
	sftp       *sftp.Client
	LastResult string //最近一次Run的结果
	authMode   string //认证方式
	publicKey  ssh.AuthMethod
}

// 创建命令行对象
// @param ip IP地址
// @param username 用户名
// @param password 密码
// @param port 端口号,默认22
func New(authMode, ip, username, password, publicKey string, port ...int) *Cli {
	cli := new(Cli)
	cli.IP = ip
	cli.Username = username
	cli.authMode = authMode
	cli.Password = password
	if authMode != "1" {
		cli.GetPublicKey(publicKey)
	}
	if len(port) <= 0 {
		cli.Port = 22
	} else {
		cli.Port = port[0]
	}
	return cli
}

func (c *Cli) Close() error {
	c.sftp.Close()
	return c.client.Close()
}

// 连接
func (c *Cli) GetPublicKey(keyPath string) error {
	var key []byte
	var err error
	if strings.HasPrefix(keyPath, "/") {
		key, err = ioutil.ReadFile(keyPath)
		if err != nil {
			return err
		}
	} else {
		key = []byte(keyPath)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return err
	}
	c.publicKey = ssh.PublicKeys(signer)
	return nil
}

// 连接
func (c *Cli) connect() error {
	defer func() {
		recover()
	}()

	config := ssh.ClientConfig{
		User:            c.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	if c.authMode == "1" {
		config.Auth = []ssh.AuthMethod{ssh.Password(c.Password)}
	} else {
		config.Auth = []ssh.AuthMethod{c.publicKey}
	}
	addr := fmt.Sprintf("%s:%d", c.IP, c.Port)

	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return err
	}
	c.client = sshClient
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	c.sftp = sftpClient
	return nil
}

// 执行shell
// @param shell shell脚本命令
func (c *Cli) Run(shell string) (string, error) {
	defer func() {
		recover()
	}()
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
func (c *Cli) NewTerminal() (*Terminal, error) {
	defer func() {
		recover()
	}()
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
	terminal.Write("echo $$")
	data, err := terminal.ReadString('\n')
	if err != nil {
		return nil, err
	}
	terminal.pid = strings.TrimSpace(data)
	location, _ := time.LoadLocation("Asia/Shanghai")
	terminal.Now = time.Now().In(location)
	terminal.IsRun = false
	go func() {
		ticker := time.NewTicker(time.Minute * 1)
		for terminal.IsRun {
			<-ticker.C
			duration := time.Now().In(location).Sub(terminal.Now)
			if duration.Minutes() > 3 {
				terminal.CloseAll()
			}
		}
	}()
	return terminal, nil
}

type Terminal struct {
	cli     *Cli
	pid     string
	session *ssh.Session
	input   io.WriteCloser
	output  *bufio.Reader
	Now     time.Time
	IsRun   bool
}

func (t *Terminal) Write(shell string) {
	defer func() {
		recover()
	}()
	t.Now = time.Now()
	t.input.Write([]byte(shell + ";echo sc-finish:$?;\n"))
}

func (t *Terminal) ReadString(delim byte, callback ...func(ip, data string)) (string, error) {
	t.Now = time.Now()
	defer func() {
		recover()
	}()
	var state string
	buf := bytes.Buffer{}
	for {
		line, err := t.output.ReadString('\n')
		if err != nil || strings.HasPrefix(strings.TrimSpace(line), "sc-finish:") {
			state = strings.TrimSpace(line)
			break
		}
		if callback != nil {
			callback[0](t.cli.IP, line)
		}
		state = strings.TrimSpace(state)
		buf.WriteString(strings.TrimSpace(line))
		buf.WriteByte('\n')
	}
	if state != "sc-finish:0" {
		return buf.String(), errors.New("命令执行失败")
	}
	return buf.String(), nil
}

// 关闭当前终端
func (t *Terminal) Close() error {
	err := t.input.Close()
	if err != nil && err != io.EOF {
		fmt.Println("ssh close error:", err)
	}
	err = t.session.Close()
	if err != nil && err != io.EOF {
		fmt.Println("ssh close error:", err)
	}
	return err
}

// 关闭当前终端以及子进程
func (t *Terminal) CloseAll() error {
	defer func() {
		recover()
	}()
	_, err := t.cli.Run("kill -9 -" + t.pid)
	if err != nil && err != io.EOF {
		fmt.Println("ssh close error:", err)
	}
	err = t.input.Close()
	if err != nil && err != io.EOF {
		fmt.Println("ssh close error:", err)
	}
	err = t.session.Close()
	if err != nil && err != io.EOF {
		fmt.Println("ssh close error:", err)
	}
	err = t.cli.Close()
	if err != nil && err != io.EOF {
		fmt.Println("ssh close error:", err)
	}
	return err
}

func (t *Terminal) UploadFile(localFilePath string, remotePath string) (string, error) {
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	var remoteFileName = path.Base(localFilePath)

	dstFile, err := t.cli.sftp.Create(path.Join(remotePath, remoteFileName))
	if err != nil {
		return "", err

	}
	defer dstFile.Close()

	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		return "", err
	}
	fmt.Println("上传中...")
	dstFile.Write(ff)
	return localFilePath + ">成功", nil
}

func (t *Terminal) UploadDirectory(localPath string, remotePath string) error {
	localFiles, err := ioutil.ReadDir(localPath)
	if err != nil {
		return err
	}
	var msg string
	for _, backupDir := range localFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			t.cli.sftp.Mkdir(remoteFilePath)
			err = t.UploadDirectory(localFilePath, remoteFilePath)
		} else {
			msg, err = t.UploadFile(path.Join(localPath, backupDir.Name()), remotePath)
			if err == nil {
				fmt.Println(msg)
			}
		}
	}
	return err
}
