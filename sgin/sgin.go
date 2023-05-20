package sgin

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/androidsr/sc-go/model"
	"github.com/androidsr/sc-go/scan"
	"github.com/androidsr/sc-go/sjwt"
	"github.com/androidsr/sc-go/syaml"
	"github.com/gin-gonic/gin"
	"github.com/timandy/routine"
)

var (
	ctrls       []interface{}
	config      *syaml.GinInfo
	threadLocal = routine.NewInheritableThreadLocal()
)

type SGin struct {
	*gin.Engine
	docs map[string]map[string]string
}

func New(cfg *syaml.GinInfo) *SGin {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	config = cfg
	router := &SGin{gin.New(), nil}
	router.docs = scan.ScanFunc(config.Scan.Pkg, config.Scan.Filter)
	return router
}

func AddRouter(ctrl ...interface{}) {
	ctrls = append(ctrls, ctrl...)
}

func (m *SGin) RunServer() error {
	m.autoRegister()
	return m.Run(fmt.Sprintf(":%d", config.Port))
}

func (s *SGin) GetContent() routine.ThreadLocal {
	return threadLocal
}

func (g *SGin) autoRegister() {
	fmt.Printf("路由注册大小：%d\n", len(ctrls))
	for _, ctrl := range ctrls {
		var value reflect.Value
		if reflect.TypeOf(ctrl).Kind() == reflect.Ptr {
			value = reflect.ValueOf(ctrl).Elem()
		} else {
			value = reflect.ValueOf(ctrl)
		}
		doc := g.docs[value.Type().Name()]
		for i := 0; i < value.NumMethod(); i++ {
			method := value.Type().Method(i)
			comment := doc[method.Name]
			if comment == "" {
				continue
			}
			cms := strings.Split(comment, " ")
			if len(cms) < 3 {
				continue
			}
			httpMethod := strings.ReplaceAll(strings.TrimSpace(cms[1]), "[", "")
			httpMethod = strings.ReplaceAll(httpMethod, "]", "")
			httpMethod = strings.ToUpper(httpMethod)
			relativePath := strings.ReplaceAll(strings.TrimSpace(cms[2]), "[", "")
			relativePath = strings.ReplaceAll(relativePath, "]", "")
			resultType := "[json]"
			if len(cms) >= 4 {
				resultType = strings.ToLower(strings.TrimSpace(cms[3]))
			}
			m := value.MethodByName(method.Name)
			fmt.Printf("路由注册：%s：%s", httpMethod, relativePath)
			g.Handle(httpMethod, relativePath, func(c *gin.Context) {
				threadLocal.Set(c)
				defer threadLocal.Remove()
				num := m.Type().NumIn()
				args := make([]reflect.Value, num)
				args[0] = reflect.ValueOf(c)
				if num == 2 {
					t := m.Type().In(1)
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					} else {
						c.JSON(http.StatusOK, model.NewFail(5000, "接收参数必需是指针类型"))
						return
					}
					data := reflect.New(t).Interface()
					err := c.ShouldBind(&data)
					if err != nil {
						c.JSON(http.StatusBadRequest, model.NewFail(400, err.Error()))
						return
					}
					args[1] = reflect.ValueOf(data)
				}
				result := m.Call(args)
				if len(result) > 0 {
					if len(result) == 2 {
						if !result[1].IsZero() {
							err := result[1].Interface().(error)
							if err != nil {
								c.JSON(http.StatusOK, model.NewFail(5000, err.Error()))
								return
							}
						}
					}
					data := result[0].Interface()
					v, ok := data.(model.HttpResult)
					if ok {
						switch resultType {
						case "[json]":
							c.JSON(http.StatusOK, v)
						case "[string]":
							c.String(http.StatusOK, "%s", data)
						}
					} else {
						switch resultType {
						case "[json]":
							c.JSON(http.StatusOK, model.NewOK(data))
						case "[string]":
							c.String(http.StatusOK, "%s", data)
						}
					}
				}
			})
		}
	}
	g.docs = nil
	ctrls = nil
}

// 跨域处理
func (m *SGin) Cors() {
	m.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
}

// 设置默认认证
func (m *SGin) WebToken(config *syaml.WebTokenInfo) {
	sjwt.New(config)
	m.Use(sjwt.JWTAuthMiddleware())
}
