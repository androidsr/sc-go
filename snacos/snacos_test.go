package snacos

import (
	"fmt"
	"testing"
	"time"

	"github.com/androidsr/paas-go/syaml"
)

func Test_GetConfig(t *testing.T) {
	configs, _ := syaml.LoadFile[syaml.PaasRoot]("../paas.yaml")
	New(configs.Paas.Nacos)
	fmt.Println(ConfigClient.GetDefaultConfig(func(namespace, group, dataId, data string) {
		fmt.Println(namespace, group, dataId, data)
	}))
	fmt.Println("-------------------")
	instance, err := NamingClient.GetInstance("paas-go")
	if err != nil {
		panic(err)
	}
	fmt.Println(instance.Ip, instance.Port)
	time.Sleep(time.Second * 10)
}
