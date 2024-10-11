package snacos

import (
	"sc-go/syaml"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var (
	config *syaml.NacosInfo
	param  vo.NacosClientParam
)

func New(cfg *syaml.NacosInfo) {
	config = cfg
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(config.IpAddr, config.Port, constant.WithContextPath("/nacos")),
	}
	cc := constant.NewClientConfig(
		constant.WithNamespaceId(config.Config.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
	)
	param = vo.NacosClientParam{
		ClientConfig:  cc,
		ServerConfigs: sc,
	}
	NewNacosConfig()
	NewNacosNaming()
}
