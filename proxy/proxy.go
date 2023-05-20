package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/androidsr/paas-go/syaml"
)

var (
	config *syaml.ProxyInfo
)

func New(config *syaml.ProxyInfo) {
	for _, v := range config.Web {
		http.Handle(v.Path, http.FileServer(http.Dir(v.Dir)))
	}
	for _, v := range config.Server {
		fmt.Println(v.Name, v.Addr)
		proxyServer(v.Name, v.Addr)
	}
	if config.Cert == "" || config.Key == "" {
		http.ListenAndServe(":"+config.Port, nil)
	} else {
		http.ListenAndServeTLS(":"+config.Port, config.Cert, config.Key, nil)
	}
}

func proxyServer(pattern string, addr string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		target := r.RequestURI
		var key string
		i := strings.Index(target[1:], "/")
		key = target[1 : i+1]
		if addr == "" {
			return
		}
		fmt.Println(target, key)
		last := strings.Index(target, "?")
		var curl string
		if last != -1 {
			curl = target[7+len(key) : last]
		} else {
			curl = target[7+len(key):]
		}
		fmt.Println(curl)
		proxy := httputil.ReverseProxy{Director: func(req *http.Request) {
			n := strings.Index(addr, ":")
			req.URL.Scheme = addr[:n]
			ip := addr[n+3:]
			n = strings.Index(ip, "/")
			path := ""
			fmt.Println(ip)
			if n != -1 {
				path = ip[n:]
				ip = ip[:n]
			}
			fmt.Println(n, ip, path)
			req.URL.Host = ip
			req.URL.Path = path + curl
		}}
		proxy.ServeHTTP(w, r)
	})
}
