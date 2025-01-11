package mapper

import (
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
	db *gorm.DB
)

type Mapper[T any] struct {
	*gorm.DB
}

// New 初始化数据库连接
func Initdb(config *syaml.GormInfo) *gorm.DB {
	var dialector gorm.Dialector
	// 根据配置选择对应的数据库驱动
	switch config.Driver {
	case "mysql":
		dialector = mysql.Open(config.Url)
	case "postgres":
		dialector = postgres.Open(config.Url)
	case "sqlite":
		dialector = sqlite.Open(config.Url)
	}
	// 配置日志
	var showLog logger.Interface
	if config.ShowSql {
		showLog = logger.Default.LogMode(logger.Info)
	}
	// 初始化数据库连接
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
	// 配置数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("数据库初始化失败:%s", err.Error())
		return nil
	}
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	return db
}

func NewHelper[T any]() *Mapper[T] {
	return &Mapper[T]{db}
}

func NewTransaction[T any](db *gorm.DB) *Mapper[T] {
	return &Mapper[T]{db}
}

// Exists 判断记录是否存在
func (m *Mapper[T]) Exists(query *T) bool {
	return m.GetCount(query) > 0
}

// GetCount 获取记录数
func (m *Mapper[T]) GetCount(query *T) int64 {
	var count int64
	// 通过条件查询记录数
	if err := m.DB.Model(query).Where(query).Count(&count).Error; err != nil {
		log.Printf("GetCount Error: %v", err)
		return 0
	}
	return count
}

// SelectCount 根据SQL查询记录数
func (m *Mapper[T]) SelectCount(sql string, values ...interface{}) int64 {
	var count int64
	sql = fmt.Sprintf("select count(*) from (%s) t", sql)
	// 执行SQL并获取记录数
	if err := m.DB.Raw(sql, values...).Scan(&count).Error; err != nil {
		log.Printf("SelectCount Error: %v", err)
		return 0
	}
	return count
}

// Insert 插入数据
func (m *Mapper[T]) Insert(value *T) error {
	// 创建记录
	return m.DB.Create(value).Error
}

// InsertBatch 批量插入数据
func (m *Mapper[T]) InsertBatch(values *[]T) error {
	// 批量创建记录
	return m.DB.CreateInBatches(values, 300).Error
}

// Tx 使用事务执行操作
func (m *Mapper[T]) Tx(fc func(tx *gorm.DB) error) error {
	return m.DB.Transaction(fc)
}

// SaveOrUpdate 保存或更新记录
func (m *Mapper[T]) SaveOrUpdate(value *T) *gorm.DB {
	return m.DB.Save(value)
}

// Update 更新记录
func (m *Mapper[T]) Update(value *T, query string, args ...interface{}) *gorm.DB {
	return m.DB.Where(query, args...).Updates(value)
}

// Delete 删除记录，使用ID或其他条件进行删除
func (m *Mapper[T]) DeleteById(ids ...interface{}) *gorm.DB {
	return m.DB.Model(new(T)).Delete(new(T), ids)
}

// Delete 删除记录，使用其他条件进行删除
func (m *Mapper[T]) Delete(query string, values ...interface{}) *gorm.DB {
	return m.DB.Model(new(T)).Where(query, values...).Delete(new(T))
}

// Delete 删除记录，使用ID或其他条件进行删除
func (m *Mapper[T]) Delete2(obj *T) *gorm.DB {
	return m.DB.Model(new(T)).Where(obj).Delete(new(T))
}

// SelectList 查询列表
func (m *Mapper[T]) SelectList(query *T) ([]T, error) {
	result := make([]T, 0)
	err := m.DB.Model(new(T)).Where(query).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Mapper[T]) SelectList2(query string, values ...interface{}) ([]T, error) {
	result := make([]T, 0)
	err := m.DB.Model(new(T)).Where(query, values...).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Mapper[T]) SelectList3(result interface{}, query string, values ...interface{}) error {
	err := m.DB.Model(new(T)).Where(query).Find(&result).Error
	if err != nil {
		return err
	}
	return nil
}

func (m *Mapper[T]) SelectList4(result interface{}, query *T) error {
	err := m.DB.Model(new(T)).Where(query).Find(&result).Error
	if err != nil {
		return err
	}
	return nil
}

// SelectAll 查询所有记录
func (m *Mapper[T]) SelectAll() ([]T, error) {
	data := make([]T, 0)
	err := m.DB.Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectOne 查询单条记录
func (m *Mapper[T]) SelectOne(data *T) error {
	err := m.DB.Where(data).First(data).Error
	if err != nil {
		return err
	}
	return nil
}

// SelectSQL 执行SQL查询
func (m *Mapper[T]) SelectSQL(data interface{}, sql string, values ...interface{}) error {
	return m.DB.Raw(sql, values...).Scan(data).Error
}

// SelectPage 分页查询
func (m *Mapper[T]) SelectPage(data interface{}, page *model.PageInfo, sql string, values ...interface{}) *model.PageResult {
	result := new(model.PageResult)
	if page != nil {
		if page.Current == 0 {
			page.Current = 1
		}
		// 获取总记录数
		count := m.SelectCount(sql, values...)
		result.Current = page.Current
		result.Size = page.Size
		if count == 0 {
			return nil
		}
		result.Total = int64(count)
		offset := (page.Current - 1) * page.Size
		sql = fmt.Sprintf("select * from (%s) t LIMIT ? OFFSET ?", sql)
		values = append(values, page.Size, offset)
	}
	// 执行分页查询
	if err := m.DB.Raw(sql, values...).Scan(data).Error; err != nil {
		log.Printf("SelectPage Error: %v", err)
		return nil
	}
	result.Rows = data
	return result
}
