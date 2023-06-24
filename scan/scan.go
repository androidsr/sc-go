package scan

import (
	"bytes"
	"encoding/json"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

/**
 * 扫描指定目录下注释，按结构体:方法：注释信息生成map
 */
func ScanFunc(dir string, preFilter string) map[string]map[string]string {
	result := make(map[string]map[string]string, 0)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		_, err = os.Stat("sc-go-router")
		if os.IsNotExist(err) {
			panic("路由配置文件不存在")
		}
		bs, err := os.ReadFile("sc-go-router")
		if err != nil {
			panic("路由配置文件读取失败")
		}
		err = json.Unmarshal(bs, &result)
		if err != nil {
			panic("解析路由配置文件失败")
		}
	} else {
		fset := token.NewFileSet()
		pkgs, _ := parser.ParseDir(fset, dir, nil, parser.ParseComments)
		for _, pkg := range pkgs {
			docPkg := doc.New(pkg, ".", doc.AllMethods)
			for _, t := range docPkg.Types {
				item := make(map[string]string, 0)
				for _, method := range t.Methods {
					doc := method.Doc
					if preFilter != "" {
						if strings.Contains(doc, preFilter) {
							buf := bytes.NewBufferString(doc)
							data := bytes.Buffer{}
							for {
								line, err := buf.ReadString('\n')
								if err != nil {
									break
								}
								if strings.HasPrefix(line, preFilter) {
									data.WriteString(line)
								}
							}
							item[method.Name] = data.String()
						}
					} else {
						item[method.Name] = doc
					}
				}
				result[t.Name] = item
			}
		}
		bs, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		os.WriteFile("sc-go-router", bs, 0666)
	}
	return result
}
