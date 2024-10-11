package snacos

import (
	"fmt"
	"log"

	"sc-go/sc"
	"sc-go/syaml"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var (
	NamingClient *NacosNaming
)

func NewNacosNaming() *NacosNaming {
	if NamingClient == nil {
		NamingClient = new(NacosNaming)
		NamingClient.config = config
		ip := NamingClient.config.Discovery.Ip
		if ip == "" {
			NamingClient.ip = sc.GetIP(NamingClient.config.Discovery.Prefix)
		} else {
			NamingClient.ip = ip
		}
		var err error
		NamingClient.client, err = clients.NewNamingClient(param)
		if err != nil {
			panic(err)
		}
		err = NamingClient.Register()
		if err != nil {
			panic(err)
		}
	}
	return NamingClient
}

type NacosNaming struct {
	config *syaml.NacosInfo
	client naming_client.INamingClient
	ip     string
}

// 注册服务
func (m *NacosNaming) Register() error {
	discovery := m.config.Discovery
	if discovery.Ip == "" {
		discovery.Ip = m.ip
	}
	_, err := m.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          discovery.Ip,
		Port:        discovery.Port,
		GroupName:   discovery.Group,
		ServiceName: discovery.ServiceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	})
	return err
}

// 注销实例
func (m *NacosNaming) Deregister() error {
	discovery := m.config.Discovery
	if discovery.Ip == "" {
		discovery.Ip = m.ip
	}

	m.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          discovery.Ip,
		Port:        discovery.Port,
		GroupName:   discovery.Group,
		ServiceName: discovery.ServiceName,
		Ephemeral:   true,
	})
	return nil
}

// 获取服务
func (m *NacosNaming) GetService(serviceName string) (model.Service, error) {
	discovery := m.config.Discovery
	return m.client.GetService(vo.GetServiceParam{
		ServiceName: serviceName,
		GroupName:   discovery.Group,
	})

}

// 获取实例
func (m *NacosNaming) GetInstance(serviceName string) (*model.Instance, error) {
	discovery := m.config.Discovery
	return m.client.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   discovery.Group,
	})
}

// 通过服务名获取url地址
func (m *NacosNaming) GetUrl(serviceName string, path string) string {
	instance, err := NamingClient.GetInstance(serviceName)
	if err != nil {
		log.Println(err)
		return ""
	}
	return fmt.Sprintf("http://%s:%d/%s", instance.Ip, instance.Port, path)
}
