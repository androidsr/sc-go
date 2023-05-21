package sorm

import (
	"bytes"
	"database/sql"
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
	DB       *Sorm
	autoFill map[string]FillFunc
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
	autoFill = make(map[string]FillFunc, 0)
	pSqlx := &Sorm{db, config}
	return pSqlx
}

// 自动填充处理
type FillFunc func() any

// 增加字段进行自动填充
func AddAutoFill(column string, call FillFunc) {
	autoFill[column] = call
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
	tableModel := GetField(obj, false)
	condi := setKeyword(tableModel)
	sql := fmt.Sprintf("select count(*) from %s where 1=1 %s", tableModel.TableName, condi)
	var count int
	err := m.DB.Get(&count, sql, tableModel.values...)
	if err != nil {
		log.Printf("SQL error:%s\n %v", sql, err)
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
		log.Printf("SQL error:%s\n %v", sql, err)
		return 0
	}
	return count
}

// 插入数据
func (m *Sorm) Insert(obj interface{}) int64 {
	tableModel := GetField(obj, true)
	sql := insertSQL(tableModel)
	printSQL(sql, tableModel.values...)
	ret, err := m.DB.Exec(sql, tableModel.values...)
	return getAffectedRow(ret, err)
}

// 插入数据（同一事物db）
func (m *Sorm) InsertTx(db *sqlx.Tx, obj interface{}) int64 {
	tableModel := GetField(obj, true)
	sql := insertSQL(tableModel)
	printSQL(sql, tableModel.values...)
	ret, err := db.Exec(sql, tableModel.values...)
	return getAffectedRow(ret, err)
}

func insertSQL(tableModel *ModelInfo) string {
	vals := make([]string, 0)
	columns := make([]string, 0)
	for i := 0; i < len(tableModel.tags); i++ {
		vals = append(vals, "?")
		columns = append(columns, tableModel.tags[i].Column)
	}
	sql := fmt.Sprintf("insert into %s(%s) values (%s)", tableModel.TableName, strings.Join(columns, ", "), strings.Join(vals, ", "))
	return sql
}

// 按ID更新非空字段
func (m *Sorm) UpdateById(obj interface{}) int64 {
	tableModel := GetField(obj, true)
	sql := updateSQL(tableModel, tableModel.PrimaryKey)
	printSQL(sql, tableModel.values...)
	ret, err := m.DB.Exec(sql, tableModel.values...)
	return getAffectedRow(ret, err)
}

// 更新数据（指定条件列）
func (m *Sorm) Update(obj interface{}, condition ...string) int64 {
	if len(condition) == 0 {
		log.Fatal("更新语句条件为空")
		return 0
	}
	tableModel := GetField(obj, true)
	sql := updateSQL(tableModel, condition...)
	printSQL(sql, tableModel.values...)
	ret, err := m.DB.Exec(sql, tableModel.values...)
	return getAffectedRow(ret, err)
}

// 更新数据（指定条件列，同一事物db）
func (m *Sorm) UpdateTx(db *sqlx.Tx, obj interface{}, condition ...string) int64 {
	if len(condition) == 0 {
		log.Fatal("更新语句条件为空")
		return 0
	}
	tableModel := GetField(obj, true)
	sql := updateSQL(tableModel, condition...)
	printSQL(sql, tableModel.values...)
	ret, err := db.Exec(sql, tableModel.values...)
	return getAffectedRow(ret, err)
}

func updateSQL(tableModel *ModelInfo, condition ...string) string {
	sets := bytes.Buffer{}
	conds := bytes.Buffer{}
	setValues := make([]interface{}, 0)
	condValues := make([]interface{}, 0)
	for i, tag := range tableModel.tags {
		if sc.Contains(condition, tag.Column) {
			conds.WriteString(fmt.Sprintf(" and %s = ?", tag.Column))
			condValues = append(condValues, tableModel.values[i])
		} else {
			sets.WriteString(fmt.Sprintf(" %s = ?,", tag.Column))
			setValues = append(setValues, tableModel.values[i])
		}
	}
	sql := fmt.Sprintf("update %s set %s where 1=1 %s", tableModel.TableName, sets.String()[:sets.Len()-1], conds.String())
	tableModel.values = append(setValues, condValues...)
	return sql
}

// 删除数据
func (m *Sorm) Delete(obj interface{}) int64 {
	tableModel := GetField(obj, false)
	sql := deleteSQL(tableModel)
	printSQL(sql, tableModel.values...)
	ret, err := m.DB.Exec(sql, tableModel.values...)
	return getAffectedRow(ret, err)
}

// 删除数据（同一事务db）
func (m *Sorm) DeleteTx(db *sqlx.Tx, obj interface{}) int64 {
	tableModel := GetField(obj, false)
	sql := deleteSQL(tableModel)
	printSQL(sql, tableModel.values...)
	ret, err := db.Exec(sql, tableModel.values...)
	return getAffectedRow(ret, err)
}

func deleteSQL(tableModel *ModelInfo) string {
	condition := bytes.Buffer{}
	for _, tag := range tableModel.tags {
		condition.WriteString(fmt.Sprintf(" and %s = ? ", tag.Column))
	}
	sql := fmt.Sprintf("delete from %s where 1=1 %s", tableModel.TableName, condition.String())
	return sql
}

// 分页查询数据
func (m *Sorm) SelectPage(data interface{}, sql string, page model.PageInfo, query interface{}) *model.PageResult {
	tableModel := GetField(query, false)
	condi := setKeyword(tableModel)
	sql = fmt.Sprintf("%s %s", sql, condi)

	result := new(model.PageResult)
	if &page != nil {
		if page.Current == 0 {
			page.Current = 1
		}
		count := m.SelectCount(sql, tableModel.values...)
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
			for _, v := range page.Orders {
				orderBy.WriteString(fmt.Sprintf(" %s ", v.Column))
				if v.Asc {
					orderBy.WriteString("asc")
				} else {
					orderBy.WriteString("desc")
				}
			}
		}
		sql = fmt.Sprintf("select * from (%s) t %s LIMIT ? OFFSET ?", sql, orderBy.String())
		tableModel.values = append(tableModel.values, page.Size, offset)
	}
	printSQL(sql, tableModel.values...)
	err := m.DB.Select(data, sql, tableModel.values...)
	if err != nil {
		log.Printf("sql error: %v\n", err)
		return nil
	}
	result.Rows = data
	return result
}

// 分页查询数据
func (m *Sorm) SelectListPage(data interface{}, sql string, page model.PageInfo, condition []string, args ...interface{}) *model.PageResult {
	condi := bytes.Buffer{}
	values := make([]interface{}, 0)
	for i, value := range args {
		if value == nil || value == "" {
			continue
		}
		vs := sc.AssertSliceType(value)
		if len(vs) == 0 {
			continue
		}
		if condition != nil && len(condition) != 0 {
			w := condition[i]
			w = strings.ReplaceAll(w, "?", Placeholders(len(vs)))
			condi.WriteString(w)
			condi.WriteString(" ")
		}
		values = append(values, vs...)
	}
	sql = fmt.Sprintf("%s %s", sql, condi.String())

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
			for _, v := range page.Orders {
				orderBy.WriteString(fmt.Sprintf(" %s ", v.Column))
				if v.Asc {
					orderBy.WriteString("asc")
				} else {
					orderBy.WriteString("desc")
				}
			}
		}
		sql = fmt.Sprintf("select * from (%s) t %s LIMIT ? OFFSET ?", sql, orderBy.String())
		args = append(args, page.Size, offset)
	}
	printSQL(sql, values...)
	err := m.DB.Select(data, sql, values...)
	if err != nil {
		log.Printf("sql error: %v\n", err)
		return nil
	}
	result.Rows = data
	return result
}

// 分页查询数据
func (m *Sorm) Select(data interface{}, sql string, condition []string, args ...interface{}) error {
	condi := bytes.Buffer{}
	values := make([]interface{}, 0)
	for i, value := range args {
		if value == nil || value == "" {
			continue
		}
		vs := sc.AssertSliceType(value)
		if len(vs) == 0 {
			continue
		}
		if condition != nil && len(condition) != 0 {
			w := condition[i]
			w = strings.ReplaceAll(w, "?", Placeholders(len(vs)))
			condi.WriteString(w)
			condi.WriteString(" ")
		}
		values = append(values, vs...)
	}
	sql = fmt.Sprintf("%s %s", sql, condi.String())

	printSQL(sql, values...)
	err := m.DB.Select(data, sql, values...)
	if err != nil {
		log.Printf("sql error: %v\n", err)
		return err
	}
	return nil
}

// 查询集合
func (m *Sorm) SelectList(data interface{}, query interface{}, columns ...string) error {
	tableModel := GetField(query, false)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, tableModel.TableName)
	condi := setKeyword(tableModel)
	sql += condi
	err := m.DB.Select(data, sql, tableModel.values...)
	if err != nil {
		log.Printf("sql error:%v\n", err)
		return err
	}
	return nil
}

// 查询集合
func (m *Sorm) FindList(data interface{}, condition []string, args ...interface{}) error {
	condi := bytes.Buffer{}
	values := make([]interface{}, 0)
	for i, value := range args {
		if value == nil || value == "" {
			continue
		}
		vs := sc.AssertSliceType(value)
		if len(vs) == 0 {
			continue
		}
		if condition != nil && len(condition) != 0 {
			w := condition[i]
			w = strings.ReplaceAll(w, "?", Placeholders(len(vs)))
			condi.WriteString(w)
			condi.WriteString(" ")
		}
		values = append(values, vs...)
	}

	tableModel := GetField(data, false)
	sql := fmt.Sprintf("select * from %s where 1=1 ", tableModel.TableName)
	sql += condi.String()
	err := m.DB.Select(data, sql, values...)
	if err != nil {
		log.Printf("sql error:%v\n", err)
		return err
	}
	return nil
}

// 查询集合
func (m *Sorm) FindOne(data interface{}, condition []string, args ...interface{}) error {
	condi := bytes.Buffer{}
	values := make([]interface{}, 0)
	for i, value := range args {
		if value == nil || value == "" {
			continue
		}
		vs := sc.AssertSliceType(value)
		if len(vs) == 0 {
			continue
		}
		if condition != nil && len(condition) != 0 {
			w := condition[i]
			w = strings.ReplaceAll(w, "?", Placeholders(len(vs)))
			condi.WriteString(w)
			condi.WriteString(" ")
		}
		values = append(values, vs...)
	}

	tableModel := GetField(data, false)
	sql := fmt.Sprintf("select * from %s where 1=1 ", tableModel.TableName)
	sql += condi.String()
	err := m.DB.Get(data, sql, values...)
	if err != nil {
		log.Printf("sql error:%v\n", err)
		return err
	}
	return nil
}

// 查询集合
func (m *Sorm) SelectListTx(tx *sqlx.Tx, data interface{}, query interface{}, columns ...string) error {
	tableModel := GetField(query, false)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, tableModel.TableName)
	condi := setKeyword(tableModel)
	sql += condi
	err := tx.Select(data, sql, tableModel.values...)
	if err != nil {
		log.Printf("sql error:%v\n", err)
		return err
	}
	return nil
}

// 查询一条记录
func (m *Sorm) SelectOne(data interface{}, query interface{}, columns ...string) error {
	tableModel := GetField(query, false)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, tableModel.TableName)
	condi := setKeyword(tableModel)
	sql += condi
	printSQL(sql, tableModel.values...)
	err := m.DB.Get(data, sql, tableModel.values...)
	if err != nil {
		log.Printf("sql error:%v\n", err)
		return err
	}
	return nil
}

// 查询一条记录
func (m *Sorm) GetOne(data interface{}, columns ...string) error {
	tableModel := GetField(data, false)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, tableModel.TableName)
	condi := setKeyword(tableModel)
	sql += condi
	printSQL(sql, tableModel.values...)
	err := m.DB.Get(data, sql, tableModel.values...)
	if err != nil {
		log.Printf("sql error:%v\n", err)
		return err
	}
	return nil
}

// 查询一条记录
func (m *Sorm) SelectOneTx(tx *sqlx.Tx, data interface{}, query interface{}, columns ...string) error {
	tableModel := GetField(query, false)
	var cols string
	if len(columns) == 0 {
		cols = " * "
	} else {
		cols = strings.Join(columns, ", ")
	}
	sql := fmt.Sprintf("select %s from %s where 1=1 ", cols, tableModel.TableName)
	condi := setKeyword(tableModel)
	sql += condi
	printSQL(sql, tableModel.values...)
	err := tx.Get(data, sql, tableModel.values...)
	if err != nil {
		log.Printf("sql error:%v\n", err)
		return err
	}
	return nil
}

// 打印SQL信息
func printSQL(sql string, values ...interface{}) {
	fmt.Printf("exec sql: %s\n ", sql)
	fmt.Printf("sql input values: %v\n ", values)
}

// 获取SQL执行影响行数
func getAffectedRow(ret sql.Result, err error) int64 {
	if err != nil {
		log.Printf("SQL ERROR: %v", err)
		return 0
	}
	n, err := ret.RowsAffected()
	if err != nil {
		log.Printf("SQL ERROR: %v", err)
		return 0
	}
	return n
}
