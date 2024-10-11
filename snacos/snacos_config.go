package snacos

import (
	"fmt"
	"log"

	"github.com/androidsr/sc-go/syaml"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var (
	ConfigClient *NacosConfig
)

func NewNacosConfig() *NacosConfig {
	if ConfigClient == nil {
		ConfigClient = new(NacosConfig)
		ConfigClient.config = config
		var err error
		ConfigClient.client, err = clients.NewConfigClient(param)
		if err != nil {
			panic(err)
		}
	}
	return ConfigClient
}

type NacosConfig struct {
	config *syaml.NacosInfo
	client config_client.IConfigClient
}

// 获取配置
func (m *NacosConfig) GetDefaultConfig(onChange func(namespace, group, dataId, data string)) (string, error) {
	return m.client.GetConfig(vo.ConfigParam{
		DataId:   m.config.Config.DataId,
		Group:    m.config.Config.Group,
		OnChange: onChange,
	})
}

// 获取配置
func (m *NacosConfig) GetConfig(dataId string, group string, onChange func(namespace, group, dataId, data string)) (string, error) {
	return m.client.GetConfig(vo.ConfigParam{
		DataId:   m.config.Config.DataId,
		Group:    m.config.Config.Group,
		OnChange: onChange,
	})
}

// 通过服务名获取url地址
func (m *NacosConfig) GetService(serviceName string, path string) string {
	instance, err := NamingClient.GetInstance(serviceName)
	if err != nil {
		log.Println(err)
		return ""
	}
	return fmt.Sprintf("http://%s:%d/%s", instance.Ip, instance.Port, path)
}
