package sorm

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/androidsr/sc-go/sc"
)

func getField(t interface{}, atFill bool) *ModelInfo {
	value := reflect.ValueOf(t)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	tableModel := new(ModelInfo)
	tableModel.values = make([]interface{}, 0)
	tableModel.TableName = sc.GetUnderscore(value.Type().Name())
	tableModel.tags = make([]TagInfo, 0)

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		tagItem := TagInfo{}
		tag := value.Type().Field(i).Tag
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
			key = sc.GetUnderscore(field.Type().Name())
		}
		tagItem.Column = key
		if strings.ToLower(key) == "id" {
			tableModel.PrimaryKey = key
		}
		if field.IsZero() {
			autoFunc := autoFill[key]
			if autoFunc != nil && atFill {
				val := autoFunc()
				if val == nil {
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

func setKeyword(tableModel *ModelInfo) string {
	condition := bytes.Buffer{}
	newValue := make([]interface{}, 0)
	for i, tag := range tableModel.tags {
		keyword := tag.Keyword
		switch keyword.Type {
		case Eq:
			condition.WriteString(fmt.Sprintf(" and %s = ? ", keyword.Column))
			newValue = append(newValue, tableModel.values[i])

		case Ne:
			condition.WriteString(fmt.Sprintf(" and %s <> ? ", keyword.Column))
			newValue = append(newValue, tableModel.values[i])

		case In:
			v, ok := tableModel.values[i].([]interface{})
			if !ok {

			}
			condition.WriteString(fmt.Sprintf(" and %s in(%s) ", keyword.Column, Placeholders(len(v))))
			newValue = append(newValue, v...)

		case NotIn:
			v, ok := tableModel.values[i].([]interface{})
			if !ok {

			}
			condition.WriteString(fmt.Sprintf(" and %s not in(%s) ", keyword.Column, Placeholders(len(v))))
			newValue = append(newValue, v...)

		case Gt:
			condition.WriteString(fmt.Sprintf(" and %s > ? ", keyword.Column))
			newValue = append(newValue, tableModel.values[i])

		case Lt:
			condition.WriteString(fmt.Sprintf(" and %s < ? ", keyword.Column))
			newValue = append(newValue, tableModel.values[i])

		case Ge:
			condition.WriteString(fmt.Sprintf(" and %s >= ? ", keyword.Column))
			newValue = append(newValue, tableModel.values[i])

		case Le:
			condition.WriteString(fmt.Sprintf(" and %s <= ? ", keyword.Column))
			newValue = append(newValue, tableModel.values[i])

		case Between:
			v, ok := tableModel.values[i].(BetweenInfo)
			if ok {
				condition.WriteString(fmt.Sprintf(" and %s between ? and ? ", keyword.Column))
				newValue = append(newValue, v.Left, v.Right)
			}
		case NotBetween:
			v, ok := tableModel.values[i].(BetweenInfo)
			if ok {
				condition.WriteString(fmt.Sprintf(" and %s not between ? and ? ", keyword.Column))
				newValue = append(newValue, v.Left, v.Right)
			}
		case Like:
			condition.WriteString(fmt.Sprintf(" and %s like CONCAT('%s', ?, '%s') ", keyword.Column, "%", "%"))
			newValue = append(newValue, tableModel.values[i])

		case NotLike:
			condition.WriteString(fmt.Sprintf(" and %s not like CONCAT('%s', ?, '%s') ", keyword.Column, "%", "%"))
			newValue = append(newValue, tableModel.values[i])

		case LikeLeft:
			condition.WriteString(fmt.Sprintf(" and %s like CONCAT('%s', ?) ", keyword.Column, "%"))
			newValue = append(newValue, tableModel.values[i])

		case LikeRight:
			condition.WriteString(fmt.Sprintf(" and %s like CONCAT( ?, '%s') ", keyword.Column, "%"))
			newValue = append(newValue, tableModel.values[i])

		}
	}
	tableModel.values = newValue
	return condition.String()
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
