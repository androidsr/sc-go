package sctail

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type MonitorFile struct {
	watcher    *fsnotify.Watcher
	file       *os.File
	FilePath   string
	fileSize   int64
	fileOffset int64
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
	fi, err := m.file.Stat()
	if err != nil {
		log.Printf("获取文件信息时出错：%v\n", err)
		return err
	}
	m.fileSize = fi.Size()
	m.fileOffset = int64(0)
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return nil
			}
			fmt.Println(event.Op.String())
			if event.Op.String() == "WRITE" {
				fmt.Println(fi.Size(), m.fileSize)
				if fi.Size() > m.fileSize {
					fmt.Println(event, ok)
					newContent, err := m.readNewContent(m.fileOffset)
					if err != nil {
						log.Printf("读取新增内容时出错：%v", err)
						continue
					}
					contentHandler(newContent)
					m.fileOffset = fi.Size()
					m.fileSize = fi.Size()
				}
			}
		}
	}
}

func (m *MonitorFile) readNewContent(offset int64) (string, error) {
	_, err := m.file.Seek(offset, 0)
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(m.file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
