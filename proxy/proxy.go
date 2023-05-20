package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/androidsr/sc-go/syaml"
)

func New(config *syaml.ProxyInfo) {
	for _, v := range config.Web {
		http.Handle(v.Path, http.FileServer(http.Dir(v.Dir)))
	}

	director := func(req *http.Request) {
		addrConfig := getProxyTarget(req.URL.Path, config.Server)
		target, _ := url.Parse(addrConfig.Addr)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = strings.Replace(req.URL.Path, addrConfig.Name, "", 1)
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
	}
	// 创建HTTP服务器
	server := http.Server{
		Addr:    ":" + config.Port, // 反向代理服务器监听的地址和端口
		Handler: proxy,             // 使用反向代理处理请求
	}
	// 启动服务器
	var err error
	if config.Cert == "" || config.Key == "" {
		err = server.ListenAndServe()
	} else {
		err = server.ListenAndServeTLS(config.Cert, config.Key)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// 根据请求路径从映射中获取代理目标URL
func getProxyTarget(path string, targets []syaml.ProxyServer) syaml.ProxyServer {
	for _, target := range targets {
		if strings.HasPrefix(path, target.Name) {
			return target
		}
	}
	return syaml.ProxyServer{}
}
