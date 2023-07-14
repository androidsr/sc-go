package sbuilder

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/androidsr/sc-go/sc"
	"github.com/opentracing/opentracing-go/log"
)

const (
	Eq         = "eq"
	Ne         = "ne"
	In         = "in"
	NotIn      = "notIn"
	Gt         = "gt"
	Lt         = "lt"
	Ge         = "ge"
	Le         = "le"
	Between    = "between"
	NotBetween = "notBetween"
	Like       = "like"
	NotLike    = "notLike"
	LikeLeft   = "likeLeft"
	LikeRight  = "likeRight"
)

type SelectBuilder struct {
	Sql    bytes.Buffer
	link   string
	Values []interface{}
	links  bool
}

func StructToBuilder(obj interface{}, sql string) *SelectBuilder {
	info := GetField(obj, 0)
	_, values := info.GetDbValues(QUERY)
	if sql == "" {
		sql = fmt.Sprintf("select * from %s where 1=1 ", info.TableName)
	}
	condi := BuildQuery(info)
	sql += condi.Sql.String()
	builder := Builder(sql)
	builder.Values = values
	return builder
}

func Builder(sql string) *SelectBuilder {
	builder := new(SelectBuilder)
	builder.Sql = *bytes.NewBufferString(sql)
	builder.link = "and"
	builder.links = false
	builder.Values = make([]interface{}, 0)
	if sql != "" {
		if !strings.Contains(sql, " where ") {
			builder.Sql.WriteString(" where 1=1 ")
		}
	}
	return builder
}

func (m *SelectBuilder) Eq(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s = ? ", m.link, column)
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) Ne(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s <> ? ", m.link, column)
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) In(column string, value interface{}) string {
	if value == nil {
		return ""
	}
	v := sc.AssertSliceType(value)
	if len(v) != 0 {
		sql := fmt.Sprintf(" %s %s in(%s) ", m.link, column, Placeholders(len(v)))
		m.Values = append(m.Values, v...)
		if m.links {
			return sql
		} else {
			m.Sql.WriteString(sql)
		}
	}
	return ""
}

func (m *SelectBuilder) NotIn(column string, value interface{}) string {
	if value == nil {
		return ""
	}
	v := sc.AssertSliceType(value)
	if len(v) != 0 {
		sql := fmt.Sprintf(" %s %s not in(%s) ", m.link, column, Placeholders(len(v)))
		m.Values = append(m.Values, v...)
		if m.links {
			return sql
		} else {
			m.Sql.WriteString(sql)
		}
	}
	return ""
}

func (m *SelectBuilder) Gt(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s > ? ", m.link, column)
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) Lt(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s < ? ", m.link, column)
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) Ge(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s >= ? ", m.link, column)
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) Le(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s <= ? ", m.link, column)
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) Between(column string, value BetweenInfo) string {
	if &value == nil || value.Left == nil || value.Left == "" || value.Right == nil || value.Right == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s between ? and ? ", m.link, column)
	m.Values = append(m.Values, value.Left, value.Right)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) NotBetween(column string, value BetweenInfo) string {
	if &value == nil || value.Left == nil || value.Left == "" || value.Right == nil || value.Right == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s not between ? and ? ", m.link, column)
	m.Values = append(m.Values, value.Left, value.Right)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) Like(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s like CONCAT('%s', ?, '%s') ", m.link, column, "%", "%")
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) NotLike(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s not like CONCAT('%s', ?, '%s') ", m.link, column, "%", "%")
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) LikeLeft(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s not like CONCAT('%s', ?) ", m.link, column, "%")
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) LikeRight(column string, value interface{}) string {
	if value == nil || value == "" {
		return ""
	}
	sql := fmt.Sprintf(" %s %s not like CONCAT(?, '%s') ", m.link, column, "%")
	m.Values = append(m.Values, value)
	if m.links {
		return sql
	} else {
		m.Sql.WriteString(sql)
	}
	return ""
}

func (m *SelectBuilder) And() *SelectBuilder {
	m.link = "and"
	return m
}

func (m *SelectBuilder) Or() *SelectBuilder {
	m.link = "or"
	return m
}

func (m *SelectBuilder) Ands(sql ...string) *SelectBuilder {
	if !m.links {
		log.Error(errors.New("调用Ands方法时，需先调用Multiple方法进行多条件组装"))
	}
	if len(sql) == 0 || len(sql[0]) == 0 {
		return m
	}
	m.link = "and"
	m.Sql.WriteString(" and (")
	for i, v := range sql {
		fmt.Println("执行and")
		if v == "" {
			continue
		}
		if i == 0 {
			v = strings.Replace(v, " and ", "", 1)
			v = strings.Replace(v, " or ", "", 1)
		}
		m.Sql.WriteString(fmt.Sprintf("%s ", v))
	}
	m.Sql.WriteString(") ")
	m.link = "and"
	m.links = false
	return m
}

func (m *SelectBuilder) Multiple() *SelectBuilder {
	m.links = true
	return m
}

func (m *SelectBuilder) Ors(sql ...string) *SelectBuilder {
	if !m.links {
		log.Error(errors.New("调用Ors方法时，需先调用Multiple方法进行多条件组装"))
	}
	if len(sql) == 0 || len(sql[0]) == 0 {
		m.links = false
		return m
	}
	m.link = "or"
	m.Sql.WriteString(" or (")
	for i, v := range sql {
		fmt.Println("执行or", v)
		if v == "" {
			continue
		}
		if i == 0 {
			if i == 0 {
				v = strings.Replace(v, " and ", "", 1)
				v = strings.Replace(v, " or ", "", 1)
			}
		}
		m.Sql.WriteString(fmt.Sprintf("%s ", v))
	}
	m.Sql.WriteString(") ")
	m.link = "and"
	m.links = false
	return m
}

func (m *SelectBuilder) Append(sql string) *SelectBuilder {
	m.Sql.WriteString(" " + sql)
	return m
}

func (m *SelectBuilder) Build() (string, []interface{}) {
	return m.Sql.String(), m.Values
}
