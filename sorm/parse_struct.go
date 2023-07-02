package sorm

import (
	"reflect"
	"strings"

	"github.com/androidsr/sc-go/sc"
	"github.com/oleiade/reflections"
)

const (
	EXEC  OrmAction = 1
	QUERY OrmAction = 2
)

type OrmAction int

type StructInfo struct {
	TableName  string
	PrimaryKey string
	Fields     []FieldInfo
}

func (m *StructInfo) GetDbValues(action OrmAction) ([]string, []interface{}) {
	columns := make([]string, 0)
	values := make([]interface{}, 0)
	for _, v := range m.Fields {
		if action == EXEC {
			columns = append(columns, v.TagDB)
		} else if action == QUERY {
			columns = append(columns, v.TagColumn)
		}
		values = append(values, v.Value)
	}
	return columns, values
}

type FieldInfo struct {
	Name       string
	TagDB      string
	TagColumn  string
	TagKeyword string
	Value      interface{}
}

func GetField(obj interface{}, fillType int) *StructInfo {
	result := new(StructInfo)
	result.Fields = make([]FieldInfo, 0)
	result.TableName = sc.GetUnderscore(reflect.Indirect(reflect.ValueOf(obj)).Type().Name())

	fields, _ := reflections.Fields(obj)
	var item FieldInfo
	for _, fName := range fields {
		kind, _ := reflections.GetFieldKind(obj, fName)
		tagDB, _ := reflections.GetFieldTag(obj, fName, "db")
		tagKeyword, _ := reflections.GetFieldTag(obj, fName, "keyword")
		tagColumn, _ := reflections.GetFieldTag(obj, fName, "column")
		tagJson, _ := reflections.GetFieldTag(obj, fName, "json")
		if kind == reflect.Struct {
			value, _ := reflections.GetField(obj, fName)
			if tagDB == "-" {
				continue
			}
			if tagColumn == "-" {
				continue
			}
			pResult := GetField(value, fillType)
			result.Fields = append(result.Fields, pResult.Fields...)
		} else {
			item = FieldInfo{}
			item.Name = fName
			item.TagDB = tagDB
			item.TagKeyword = tagKeyword
			item.TagColumn = tagColumn
			if item.TagDB == "-" {
				continue
			}
			if item.TagColumn == "-" {
				continue
			}
			if item.TagDB == "" {
				if strings.Contains(tagJson, ",") {
					tagJson = strings.Split(tagJson, ",")[0]
				}
				item.TagDB = sc.GetUnderscore(tagJson)
			}
			if item.TagDB == "" {
				item.TagDB = sc.GetUnderscore(fName)
			}
			if strings.Contains(item.TagDB, ",") || strings.Contains(item.TagDB, " ") {
				var ks []string
				if strings.Contains(item.TagDB, ",") {
					ks = strings.Split(item.TagDB, ",")
				} else if strings.Contains(item.TagDB, " ") {
					ks = strings.Split(item.TagDB, " ")
				}
				item.TagDB = ks[0]
				pk := ks[1]
				if pk == "primary_key" || pk == "primaryKey" || pk == "pk" {
					result.PrimaryKey = item.TagDB
				}
			}
			if result.PrimaryKey == "" && strings.ToLower(item.TagDB) == "id" {
				result.PrimaryKey = item.TagDB
			}
			if item.TagColumn == "" {
				item.TagColumn = item.TagDB
			}
			if item.TagKeyword == "" {
				item.TagKeyword = "eq"
			}

			value, _ := reflections.GetField(obj, fName)
			if value == nil || value == "" || value == "<nil>" {
				var autoFunc FillFunc
				if fillType == 1 {
					autoFunc = insertFill[item.TagDB]
				} else if fillType == 2 {
					autoFunc = updateFill[item.TagDB]
				} else {
					continue
				}
				if autoFunc == nil {
					continue
				}
				val := autoFunc()
				if val == nil || val == "" {
					continue
				}
				value = val
				reflections.SetField(obj, fName, value)
			}
			if value == nil || value == "<nil>" {
				continue
			}
			switch value.(type) {
			case string:
				if value == "" {
					continue
				}
			}
			if value == "-" {
				value = ""
			}
			item.Value = value
			result.Fields = append(result.Fields, item)
		}
	}
	return result
}

type BetweenInfo struct {
	Left  interface{} `json:"start"`
	Right interface{} `json:"end"`
}

func buildQuery(info *StructInfo) *SelectBuilder {
	builder := Builder("")
	for _, item := range info.Fields {
		keyword := item.TagKeyword
		switch keyword {
		case Eq:
			builder.Eq(item.TagColumn, item.Value)
		case Ne:
			builder.Ne(item.TagColumn, item.Value)
		case In:
			builder.In(item.TagColumn, item.Value)
		case NotIn:
			builder.NotIn(item.TagColumn, item.Value)
		case Gt:
			builder.Gt(item.TagColumn, item.Value)
		case Lt:
			builder.Lt(item.TagColumn, item.Value)
		case Ge:
			builder.Ge(item.TagColumn, item.Value)
		case Le:
			builder.Le(item.TagColumn, item.Value)
		case Between:
			value, ok := item.Value.(BetweenInfo)
			if !ok {
				break
			}
			builder.Between(item.TagColumn, value)
		case NotBetween:
			value, ok := item.Value.(BetweenInfo)
			if !ok {
				break
			}
			builder.NotBetween(item.TagColumn, value)
		case Like:
			builder.Like(item.TagColumn, item.Value)
		case NotLike:
			builder.NotLike(item.TagColumn, item.Value)
		case LikeLeft:
			builder.LikeLeft(item.TagColumn, item.Value)
		case LikeRight:
			builder.LikeRight(item.TagColumn, item.Value)
		}
	}
	return builder
}

func Placeholders(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("?")
		if i != n-1 {
			b.WriteString(", ")
		}
	}

	return b.String()
}
