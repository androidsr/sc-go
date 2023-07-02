package sorm

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/androidsr/sc-go/model"
	"github.com/androidsr/sc-go/sc"
	"github.com/androidsr/sc-go/syaml"
	"github.com/jmoiron/sqlx"
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

var (
	DB         *Sorm
	insertFill map[string]FillFunc
	updateFill map[string]FillFunc
)

func New(config *syaml.SqlxInfo) *Sorm {
	db, err := sqlx.Open(config.Driver, config.Url)
	if err != nil {
		log.Printf("数据库初始化失败:%s", err.Error())
		return nil
	}
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetMaxOpenConns(config.MaxOpen)
	err = db.Ping()
	if err != nil {
		log.Printf("数据库连接异常：%s", err.Error())
		return nil
	}
	insertFill = make(map[string]FillFunc, 0)
	pSqlx := &Sorm{db, config}
	return pSqlx
}

// 自动填充处理
type FillFunc func() any

// 增加字段进行自动填充
func AddInsertFill(column string, call FillFunc) {
	insertFill[column] = call
}

func AddUpdateFill(column string, call FillFunc) {
	updateFill[column] = call
}

type Sorm struct {
	*sqlx.DB
	config *syaml.SqlxInfo
}

// 判断数据是否存在
func (m *Sorm) Exists(obj interface{}) bool {
	count := m.GetCount(obj)
	return count > 0
}

// 按条件获取数据条数
func (m *Sorm) GetCount(obj interface{}) int {
	info := GetField(obj, 0)
	builder := buildQuery(info)
	sql := fmt.Sprintf("select count(*) from %s where 1=1 %s", info.TableName, builder.sql.String())
	var count int
	err := m.DB.Get(&count, sql, builder.values...)
	if err != nil {
		log.Printf("执行SQL异常:%s\n %v", sql, err)
		return 0
	}
	return count
}

// 数据总条数
func (m *Sorm) SelectCount(sql string, values ...interface{}) int {
	var count int
	sql = fmt.Sprintf("select count(*) from (%s) t", sql)
	printSQL(sql, values...)
	err := m.DB.Get(&count, sql, values...)
	if err != nil {
		log.Printf("执行SQL异常:%s\n %v", sql, err)
		return 0
	}
	return count
}

// 插入数据
func (m *Sorm) Insert(obj interface{}) error {
	info := GetField(obj, 1)
	columns, values := info.GetDbValues(EXEC)
	sql := insertSQL(info.TableName, columns)
	printSQL(sql, values...)
	ret, err := m.DB.Exec(sql, values...)
	return getAffectedRow(ret, err)
}

// 插入数据（同一事物db）
func (m *Sorm) InsertTx(db *sqlx.Tx, obj interface{}) error {
	info := GetField(obj, 1)
	column, values := info.GetDbValues(EXEC)
	sql := insertSQL(info.TableName, column)
	printSQL(sql, values...)
	ret, err := db.Exec(sql, values...)
	return getAffectedRow(ret, err)
}

func insertSQL(tableName string, columns []string) string {
	vals := make([]string, 0)
	for i := 0; i < len(columns); i++ {
		vals = append(vals, "?")
	}
	sql := fmt.Sprintf("insert into %s(%s) values (%s)", tableName, strings.Join(columns, ", "), strings.Join(vals, ", "))
	return sql
}

// 按ID更新非空字段
func (m *Sorm) UpdateById(obj interface{}) error {
	info := GetField(obj, 2)
	column, values := info.GetDbValues(EXEC)
	sql, values := updateSQL(info.TableName, column, values, info.PrimaryKey)
	printSQL(sql, values...)
	ret, err := m.DB.Exec(sql, values...)
	return getAffectedRow(ret, err)
}

// 更新数据（指定条件列）
func (m *Sorm) Update(obj interface{}, condition ...string) error {
	if len(condition) == 0 {
		return errors.New("更新语句条件为空")
	}
	info := GetField(obj, 2)
	column, values := info.GetDbValues(EXEC)
	sql, values := updateSQL(info.TableName, column, values, info.PrimaryKey)
	printSQL(sql, values...)
	ret, err := m.DB.Exec(sql, values...)
	return getAffectedRow(ret, err)
}

// 更新数据（指定条件列，同一事物db）
func (m *Sorm) UpdateTx(db *sqlx.Tx, obj interface{}, condition ...string) error {
	if len(condition) == 0 {
		return errors.New("更新语句条件为空")
	}
	info := GetField(obj, 2)
	column, values := info.GetDbValues(EXEC)
	sql, values := updateSQL(info.TableName, column, values, info.PrimaryKey)
	printSQL(sql, values...)
	ret, err := db.Exec(sql, values...)
	return getAffectedRow(ret, err)
}

func updateSQL(tableName string, columns []string, values []interface{}, condition ...string) (string, []interface{}) {
	sets := bytes.Buffer{}
	conds := bytes.Buffer{}
	setValues := make([]interface{}, 0)
	condValues := make([]interface{}, 0)
	for i, column := range columns {
		if sc.Contains(condition, column) {
			conds.WriteString(fmt.Sprintf(" and %s = ?", column))
			condValues = append(condValues, values[i])
		} else {
			sets.WriteString(fmt.Sprintf(" %s = ?,", column))
			setValues = append(setValues, values[i])
		}
	}
	sql := fmt.Sprintf("update %s set %s where 1=1 %s", tableName, sets.String()[:sets.Len()-1], conds.String())
	return sql, append(setValues, condValues...)
}

// 删除数据
func (m *Sorm) Delete(obj interface{}) error {
	info := GetField(obj, 0)
	column, values := info.GetDbValues(EXEC)
	sql := deleteSQL(info.TableName, column)
	printSQL(sql, values...)
	ret, err := m.DB.Exec(sql, values...)
	return getAffectedRow(ret, err)
}

// 删除数据（同一事务db）
func (m *Sorm) DeleteTx(db *sqlx.Tx, obj interface{}) error {
	info := GetField(obj, 0)
	column, values := info.GetDbValues(EXEC)
	sql := deleteSQL(info.TableName, column)
	printSQL(sql, values...)
	ret, err := db.Exec(sql, values...)
	return getAffectedRow(ret, err)
}

func deleteSQL(tableName string, columns []string) string {
	condition := bytes.Buffer{}
	for _, column := range columns {
		condition.WriteString(fmt.Sprintf(" and %s = ? ", column))
	}
	sql := fmt.Sprintf("delete from %s where 1=1 %s", tableName, condition.String())
	return sql
}

// 分页查询数据
func (m *Sorm) SelectPage(data interface{}, page model.PageInfo, sql string, values ...interface{}) *model.PageResult {
	result := new(model.PageResult)
	if &page != nil {
		if page.Current == 0 {
			page.Current = 1
		}
		count := m.SelectCount(sql, values...)
		result.Current = page.Current
		result.Size = page.Size
		if count == 0 {
			return nil
		}
		result.Total = int64(count)
		offset := (page.Current - 1) * page.Size
		result.Current = offset
		orderBy := bytes.Buffer{}
		if page.Orders != nil {
			orderBy.WriteString("order by ")
			for i, v := range page.Orders {
				orderBy.WriteString(fmt.Sprintf(" %s ", v.Column))
				if v.Asc {
					orderBy.WriteString("asc")
				} else {
					orderBy.WriteString("desc")
				}
				if i != len(page.Orders)-1 {
					orderBy.WriteString(", ")
				}
			}
		}
		sql = fmt.Sprintf("select * from (%s) t %s LIMIT ? OFFSET ?", sql, orderBy.String())
		values = append(values, page.Size, offset)
	}
	printSQL(sql, values...)
	err := m.DB.Select(data, sql, values...)
	if err != nil {
		log.Printf("执行SQL异常: %v\n", err)
		return nil
	}
	result.Rows = data
	return result
}

// 查询集合
func (m *Sorm) SelectList(data interface{}, query interface{}, columns ...string) error {
	info := GetField(query, 0)
	_, values := info.GetDbValues(QUERY)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, info.TableName)
	condi := buildQuery(info)
	sql += condi.sql.String()
	err := m.DB.Select(data, sql, values...)
	if err != nil {
		log.Printf("执行SQL异常:%v\n", err)
		return err
	}
	return nil
}

// 查询集合
func (m *Sorm) SelectListTx(tx *sqlx.Tx, data interface{}, query interface{}, columns ...string) error {
	info := GetField(query, 0)
	_, values := info.GetDbValues(QUERY)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, info.TableName)
	condi := buildQuery(info)
	sql += condi.sql.String()
	err := tx.Select(data, sql, values...)
	if err != nil {
		log.Printf("执行SQL异常:%v\n", err)
		return err
	}
	return nil
}

// 查询一条记录
func (m *Sorm) SelectOne(data interface{}, query interface{}, columns ...string) error {
	info := GetField(query, 0)
	_, values := info.GetDbValues(QUERY)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, info.TableName)
	condi := buildQuery(info)
	sql += condi.sql.String()
	printSQL(sql, values...)
	err := m.DB.Get(data, sql, values...)
	if err != nil {
		log.Printf("执行SQL异常:%v\n", err)
		return err
	}
	return nil
}

// 查询一条记录
func (m *Sorm) GetOne(data interface{}, columns ...string) error {
	info := GetField(data, 0)
	_, values := info.GetDbValues(QUERY)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, info.TableName)
	condi := buildQuery(info)
	sql += condi.sql.String()
	printSQL(sql, values...)
	err := m.DB.Get(data, sql, values...)
	if err != nil {
		log.Printf("执行SQL异常:%v\n", err)
		return err
	}
	return nil
}

// 查询一条记录
func (m *Sorm) SelectOneTx(tx *sqlx.Tx, data interface{}, query interface{}, columns ...string) error {
	info := GetField(query, 0)
	_, values := info.GetDbValues(QUERY)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, info.TableName)
	condi := buildQuery(info)
	sql += condi.sql.String()
	printSQL(sql, values...)
	err := tx.Get(data, sql, values...)
	if err != nil {
		log.Printf("执行SQL异常:%v\n", err)
		return err
	}
	return nil
}

// 打印SQL信息
func printSQL(sql string, values ...interface{}) {
	fmt.Printf("执行SQL: %s\n%v\n", sql, values)
}

// 获取SQL执行影响行数
func getAffectedRow(ret sql.Result, err error) error {
	if err != nil {
		log.Printf("更新SQL失败: %v", err)
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		log.Printf("更新SQL失败: %v", err)
		return err
	}
	return nil
}
