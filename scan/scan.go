package scan

import (
	"bytes"
	"go/doc"
	"go/parser"
	"go/token"
	"strings"
)

/**
 * 扫描指定目录下注释，按结构体:方法：注释信息生成map
 */
func ScanFunc(dir string, preFilter string) map[string]map[string]string {
	result := make(map[string]map[string]string, 0)
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
	return result
}
