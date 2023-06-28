package sorm

import (
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx/reflectx"

	"github.com/androidsr/sc-go/sc"
)

func baseType(t reflect.Type, expected reflect.Kind) reflect.Type {
	t = reflectx.Deref(t)
	if t.Kind() != expected {
		return nil
	}
	return t
}

func GetField(t interface{}, atFill bool) *ModelInfo {
	value := reflect.Indirect(reflect.ValueOf(t))
	vType := value.Type()
	if value.Kind() == reflect.Slice {
		vType = value.Type().Elem()
		value = reflect.New(vType).Elem()
	}
	tableModel := new(ModelInfo)
	tableModel.values = make([]interface{}, 0)
	tableModel.TableName = sc.GetUnderscore(value.Type().Name())
	tableModel.tags = make([]TagInfo, 0)
	for i := 0; i < vType.NumField(); i++ {
		field := value.Field(i)
		tagItem := TagInfo{}
		tag := vType.Field(i).Tag
		key := tag.Get("db")
		if key == "" {
			key = tag.Get("json")
		} else {
			if key == "-" {
				continue
			}
			if strings.Contains(key, ",") {
				ks := strings.Split(key, ",")
				key = ks[0]
				pk := ks[1]
				if pk == "primary_key" {
					tableModel.PrimaryKey = key
				}
			}
		}
		if key == "" {
			key = sc.GetUnderscore(vType.Name())
		}
		tagItem.Column = key
		if strings.ToLower(key) == "id" {
			tableModel.PrimaryKey = key
		}
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				continue
			}
		} else if field.IsZero() {
			autoFunc := autoFill[key]
			if autoFunc != nil && atFill {
				val := autoFunc()
				if val == nil || val == "" {
					continue
				}
				field.Set(reflect.ValueOf(val))
			} else {
				continue
			}
		}

		keyword := tag.Get("keyword")
		if keyword == "" {
			keyword = "eq"
		}
		ks := strings.Split(keyword, ",")
		kTag := KeywordTag{}
		kTag.Type = ks[0]
		if len(ks) >= 2 {
			kTag.Column = ks[1]
		} else {
			kTag.Column = key
		}
		switch field.Kind() {
		case reflect.String:
			tableModel.values = append(tableModel.values, field.String())
		case reflect.Int64 | reflect.Int32:
			tableModel.values = append(tableModel.values, field.Int())
		case reflect.Float32 | reflect.Float64:
			tableModel.values = append(tableModel.values, field.Float())
		case reflect.Bool:
			tableModel.values = append(tableModel.values, field.Bool())
		case reflect.Ptr:
			tableModel.values = append(tableModel.values, field.Elem().Interface())
		default:
			continue
		}
		tagItem.Keyword = kTag
		tableModel.tags = append(tableModel.tags, tagItem)
	}
	return tableModel
}

type ModelInfo struct {
	TableName  string
	PrimaryKey string
	tags       []TagInfo
	values     []interface{}
}

type TagInfo struct {
	Column  string
	Keyword KeywordTag
}

type KeywordTag struct {
	Type   string
	Column string
}

type BetweenInfo struct {
	Left  interface{} `json:"start"`
	Right interface{} `json:"end"`
}

func buildQuery(tableModel *ModelInfo) string {
	builder := Builder("")
	for i, tag := range tableModel.tags {
		keyword := tag.Keyword
		switch keyword.Type {
		case Eq:
			builder.Eq(keyword.Column, tableModel.values[i])
		case Ne:
			builder.Ne(keyword.Column, tableModel.values[i])
		case In:
			builder.In(keyword.Column, tableModel.values[i])
		case NotIn:
			builder.NotIn(keyword.Column, tableModel.values[i])
		case Gt:
			builder.Gt(keyword.Column, tableModel.values[i])
		case Lt:
			builder.Lt(keyword.Column, tableModel.values[i])
		case Ge:
			builder.Ge(keyword.Column, tableModel.values[i])
		case Le:
			builder.Le(keyword.Column, tableModel.values[i])
		case Between:
			value, ok := tableModel.values[i].(BetweenInfo)
			if !ok {
				break
			}
			builder.Between(keyword.Column, value)
		case NotBetween:
			value, ok := tableModel.values[i].(BetweenInfo)
			if !ok {
				break
			}
			builder.NotBetween(keyword.Column, value)
		case Like:
			builder.Like(keyword.Column, tableModel.values[i])
		case NotLike:
			builder.NotLike(keyword.Column, tableModel.values[i])
		case LikeLeft:
			builder.LikeLeft(keyword.Column, tableModel.values[i])
		case LikeRight:
			builder.LikeRight(keyword.Column, tableModel.values[i])
		}
	}
	tableModel.values = builder.values
	return builder.sql.String()
}

func Placeholders(n int) string {
	var b strings.Builder
	for i := 0; i < n-1; i++ {
		b.WriteString("?,")
	}
	if n > 0 {
		b.WriteString("?")
	}
	return b.String()
}
