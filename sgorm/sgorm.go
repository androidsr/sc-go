package sgorm

import (
	"bytes"
	"fmt"
	"log"

	"github.com/androidsr/sc-go/model"
	"github.com/androidsr/sc-go/syaml"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	DB *Sgorm
)

type Sgorm struct {
	*gorm.DB
	config *syaml.GormInfo
}

func New(config *syaml.GormInfo) *Sgorm {
	var dialector gorm.Dialector
	switch config.Driver {
	case "mysql":
		dialector = mysql.Open(config.Url)
	case "postgres":
		dialector = postgres.Open(config.Url)
	case "sqlite":
		dialector = sqlite.Open(config.Url)
	}
	var showLog logger.Interface
	if config.ShowSql {
		showLog = logger.Default.LogMode(logger.Info)
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: showLog,
	})
	if err != nil {
		log.Printf("数据库初始化失败:%s", err.Error())
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("数据库初始化失败:%s", err.Error())
		return nil
	}
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	pSqlx := &Sgorm{db, config}
	return pSqlx
}

// 判断数据是否存在
func (m *Sgorm) Exists(query interface{}) bool {
	count := m.GetCount(query)
	return count > 0
}

// 按条件获取数据条数
func (m *Sgorm) GetCount(query interface{}) int64 {
	var count int64
	err := m.DB.Where(query).Count(&count).Error
	if err != nil {
		fmt.Printf("Gorm GetCount -> Error：%v\n", err)
		return 0
	}
	return count
}

// 数据总条数
func (m *Sgorm) SelectCount(sql string, values ...interface{}) int64 {
	var count int64
	sql = fmt.Sprintf("select count(*) from (%s) t", sql)
	err := m.DB.Raw(sql, values...).Scan(&count).Error
	if err != nil {
		fmt.Printf("Gorm GetCount -> Error：%v\n", err)
		return 0
	}
	return count
}

// 插入数据
func (m *Sgorm) Insert(obj interface{}) error {
	result := m.DB.Create(obj)
	return result.Error
}

// 批量插入数据
func (m *Sgorm) InsertBatch(obj []interface{}) error {
	result := m.DB.CreateInBatches(obj, 300)
	return result.Error
}
func (m *Sgorm) Tx(fc func(tx *gorm.DB) error) error {
	err := m.DB.Transaction(fc)
	return err
}

// 按ID更新非空字段
func (m *Sgorm) SaveOrUpdate(obj interface{}) *gorm.DB {
	return m.DB.Save(obj)
}

func (m *Sgorm) UpdateById(obj interface{}) *gorm.DB {
	return m.DB.Updates(obj)
}

// 更新数据（指定条件列）
func (m *Sgorm) Update(obj interface{}, query string, args ...interface{}) *gorm.DB {
	return m.DB.Where(query, args...).Updates(obj)
}

// 删除数据
func (m *Sgorm) Delete(obj interface{}, query interface{}, args ...interface{}) *gorm.DB {
	return m.DB.Where(query, args...).Delete(obj)
}

// 删除数据
func (m *Sgorm) DeleteById(obj interface{}, id interface{}) error {
	return m.DB.Delete(obj, id).Error
}

// 删除数据
func (m *Sgorm) DeleteByIds(obj interface{}, ids []interface{}) error {
	return m.DB.Delete(obj, ids).Error
}

// 查询集合
func (m *Sgorm) SelectList(data interface{}, query interface{}) error {
	return m.DB.Where(query).Find(data).Error
}

// 查询全部
func (m *Sgorm) SelectAll(data interface{}) error {
	return m.DB.Find(data).Error
}

// 查询一条记录
func (m *Sgorm) SelectOne(data interface{}, query interface{}) error {
	return m.DB.Where(query).First(data).Error
}

// 查询一条记录
func (m *Sgorm) GetOne(dataAndQuery interface{}) error {
	return m.DB.Where(dataAndQuery).First(dataAndQuery).Error
}

// 查询一条记录
func (m *Sgorm) Get(dataAndQuery interface{}) error {
	return m.DB.Where(dataAndQuery).First(dataAndQuery).Error
}

// 原生sql语句查询
func (m *Sgorm) SelectSQL(data interface{}, sql string, values ...interface{}) error {
	err := m.DB.Raw(sql, values...).Scan(data).Error
	return err
}

// 分页查询数据
func (m *Sgorm) SelectPage(data interface{}, page *model.PageInfo, sql string, values ...interface{}) *model.PageResult {
	result := new(model.PageResult)
	if page != nil {
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
	err := m.DB.Raw(sql, values...).Scan(data).Error
	if err != nil {
		log.Printf("执行SQL异常: %v\n", err)
		return nil
	}
	result.Rows = data
	return result
}
