package spoweron

import (
	"fmt"
	"os"
	"sync"

	"github.com/androidsr/sc-go/sc"
	"github.com/kardianos/service"
)

var (
	wg sync.WaitGroup
)

type Poweron func()

type program struct {
	log      service.Logger
	cfg      *service.Config
	callback Poweron
}

func (p *program) Start(s service.Service) error {
	go p.run()
	fmt.Println("启动应用程序")
	return nil
}

func (p *program) run() {
	fmt.Println("应用程序已启动")
	p.callback()
	wg.Done()
}

func (p *program) Stop(s service.Service) error {
	fmt.Println("退出应用程序")
	os.Exit(0)
	return nil
}

func Run(name string, callback Poweron) {
	err := os.Chdir(sc.GetExecuteDir())
	if err != nil {
		fmt.Println("切换当前工作目录错误：", err)
		return
	}
	wg.Add(1)
	svcConfig := &service.Config{
		Name:        name,
		DisplayName: name + " service",
		Description: name + " service for golang",
	}
	prg := &program{callback: callback}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Println(err.Error())
	}
	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			x := s.Install()
			if x != nil {
				fmt.Println("error:", x.Error())
				return
			}
			fmt.Println("服务安装成功")
			return
		} else if os.Args[1] == "uninstall" {
			x := s.Uninstall()
			if x != nil {
				fmt.Println("error:", x.Error())
				return
			}
			fmt.Println("服务卸载成功")
			return
		}
	}
	err = s.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
	wg.Wait()
}
