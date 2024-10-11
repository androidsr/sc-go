package proxy

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/androidsr/sc-go/syaml"
)

var (
	server = make(map[string]string, 0)
)

func New(config *syaml.ProxyInfo) {
	for _, v := range config.Web {
		http.Handle(v.Path, http.FileServer(http.Dir(v.Dir)))
	}

	for _, v := range config.Server {
		server[v.Name] = v.Addr
		go func(v syaml.ProxyServer) {
			if !strings.HasPrefix(v.Name, "/") {
				v.Name = "/" + v.Name
			}
			if !strings.HasSuffix(v.Name, "/") {
				v.Name = v.Name + "/"
			}
			proxy := &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					target, _ := url.Parse(v.Addr)
					req.URL.Scheme = target.Scheme
					req.URL.Host = target.Host
					if !v.Prefix {
						req.URL.Path = "/" + strings.TrimPrefix(req.URL.Path, v.Name)
					}
				},
			}
			http.HandleFunc(v.Name, func(w http.ResponseWriter, r *http.Request) {
				proxy.ServeHTTP(w, r)
			})
		}(v)
	}
	http.HandleFunc("/address", func(w http.ResponseWriter, r *http.Request) {
		bs, err := json.Marshal(server)
		if err != nil {
			w.Write([]byte("{}"))
		}
		w.Write(bs)
	})
	var err error
	if config.Cert == "" || config.Key == "" {
		err = http.ListenAndServe(":"+config.Port, nil)
	} else {
		err = http.ListenAndServeTLS(":"+config.Port, config.Cert, config.Key, nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// 获取代理地址
func GetServer(key string) string {
	return server[key]
}
