package sctail

import (
	"bufio"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type MonitorFile struct {
	watcher  *fsnotify.Watcher
	file     *os.File
	FilePath string
	readSize int64
}

func New(filePath string) *MonitorFile {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("创建监视器时出错：%v\n", err)
		return nil
	}
	err = watcher.Add(filePath)
	if err != nil {
		log.Printf("添加文件到监视器时出错：%v\n", err)
		return nil
	}
	monitor := new(MonitorFile)
	monitor.FilePath = filePath
	monitor.watcher = watcher
	return monitor
}

func (m *MonitorFile) Close() {
	m.watcher.Close()
}

func (m *MonitorFile) Start(contentHandler func(string)) error {
	var err error
	m.file, err = os.Open(m.FilePath)
	if err != nil {
		log.Printf("打开文件时出错：%v\n", err)
		return err
	}
	defer m.file.Close()
	for {
		fi, _ := m.file.Stat()
		m.readSize = fi.Size()
		event, ok := <-m.watcher.Events
		if !ok {
			return nil
		}
		if event.Op.String() == "WRITE" {
			m.readNewContent(contentHandler)
		}
	}
}

func (m *MonitorFile) readNewContent(contentHandler func(string)) {
	m.file.Seek(m.readSize, 0)
	buf := bufio.NewReader(m.file)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		contentHandler(line)
		m.readSize += int64(len(line))
	}
}
