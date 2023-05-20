package proxy

import (
	"github.com/androidsr/sc-go/sc"
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
		if sc.IsEmpty(target.Scheme) {
			target.Scheme = "http"
		}
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = strings.Replace(req.URL.Path, addrConfig.Name, "", 1)
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
	}
	// 启动服务器
	var err error
	if config.Cert == "" || config.Key == "" {
		err = http.ListenAndServe(":"+config.Port, proxy)
	} else {
		err = http.ListenAndServeTLS(":"+config.Port, config.Cert, config.Key, proxy)
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
