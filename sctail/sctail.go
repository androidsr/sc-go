package sctail

import (
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type MonitorFile struct {
	watcher    *fsnotify.Watcher
	file       *os.File
	filePath   string
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
	monitor.filePath = filePath
	return monitor
}
func (m *MonitorFile) Close() {
	m.watcher.Close()
}
func (m *MonitorFile) Start(contentHandler func(string)) error {
	var err error
	m.file, err = os.Open(m.filePath)
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
				log.Printf("监听事件出错了\n")
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				if fi.Size() > m.fileSize {
					newContent, err := m.readNewContent(m.fileOffset)
					if err != nil {
						log.Printf("读取新增内容时出错：%v", err)
					}
					contentHandler(newContent)
					m.fileOffset = fi.Size()
				}
			}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				log.Printf("监听事件出错了；%v\n", err)
				return nil
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
