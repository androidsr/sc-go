package sorm

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/androidsr/sc-go/sc"
)

type SelectBuilder struct {
	sql    bytes.Buffer
	values []interface{}
}

func Builder(sql string) *SelectBuilder {
	builder := new(SelectBuilder)
	builder.sql = *bytes.NewBufferString(sql)
	builder.values = make([]interface{}, 0)
	if strings.Contains(sql, " where ") {
		builder.sql.WriteString(" where 1=1 ")
	}
	return builder
}

func (m *SelectBuilder) Eq(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s = ? ", column))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) Ne(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s <> ? ", column))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) In(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	v := sc.AssertSliceType(value)
	if len(v) != 0 {
		m.sql.WriteString(fmt.Sprintf(" and %s in(%s) ", column, Placeholders(len(v))))
		m.values = append(m.values, v...)
	}
	return m
}

func (m *SelectBuilder) NotIn(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	v := sc.AssertSliceType(value)
	if len(v) != 0 {
		m.sql.WriteString(fmt.Sprintf(" and %s not in(%s) ", column, Placeholders(len(v))))
		m.values = append(m.values, v...)
	}
	return m
}

func (m *SelectBuilder) Gt(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s > ? ", column))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) Lt(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s < ? ", column))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) Ge(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s >= ? ", column))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) Le(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s <= ? ", column))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) Between(column string, value BetweenInfo) *SelectBuilder {
	if &value == nil || value.Left == nil || value.Left == "" || value.Right == nil || value.Right == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s between ? and ? ", column))
	m.values = append(m.values, value.Left, value.Right)
	return m
}

func (m *SelectBuilder) NotBetween(column string, value BetweenInfo) *SelectBuilder {
	if &value == nil || value.Left == nil || value.Left == "" || value.Right == nil || value.Right == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s not between ? and ? ", column))
	m.values = append(m.values, value.Left, value.Right)
	return m
}

func (m *SelectBuilder) Like(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s like CONCAT('%s', ?, '%s') ", column, "%", "%"))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) NotLike(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s not like CONCAT('%s', ?, '%s') ", column, "%", "%"))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) LikeLeft(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s not like CONCAT('%s', ?) ", column, "%"))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) LikeRight(column string, value interface{}) *SelectBuilder {
	if value == nil || value == "" {
		return m
	}
	m.sql.WriteString(fmt.Sprintf(" and %s not like CONCAT(?, '%s') ", column, "%"))
	m.values = append(m.values, value)
	return m
}

func (m *SelectBuilder) Append(column string) *SelectBuilder {
	m.sql.WriteString(" " + column)
	return m
}

func (m *SelectBuilder) Build() (string, []interface{}) {
	return m.sql.String(), m.values
}
