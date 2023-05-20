package proxy

import (
	"github.com/androidsr/sc-go/syaml"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func New(config *syaml.ProxyInfo) {
	for _, v := range config.Web {
		http.Handle(v.Path, http.FileServer(http.Dir(v.Dir)))
	}

	for _, v := range config.Server {
		if !strings.HasPrefix(v.Name, "/") {
			v.Name = "/" + v.Name
		}
		if !strings.HasSuffix(v.Name, "/") {
			v.Name = v.Name + "/"
		}
		http.HandleFunc(v.Name, func(w http.ResponseWriter, r *http.Request) {
			proxy := &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					target, _ := url.Parse(v.Addr)
					req.URL.Scheme = target.Scheme
					req.URL.Host = target.Host
					req.URL.Path = strings.Replace(req.URL.Path, v.Name, "/", 1)
				},
			}
			proxy.ServeHTTP(w, r)
		})
	}

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
