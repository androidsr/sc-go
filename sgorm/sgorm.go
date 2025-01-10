package sgorm

import (
	"fmt"
	"log"
	"time"

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

// New 初始化数据库连接
func New(config *syaml.GormInfo) *Sgorm {
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
	return &Sgorm{db, config}
}

// Exists 判断记录是否存在
func (m *Sgorm) Exists(query interface{}) bool {
	return m.GetCount(query) > 0
}

// GetCount 获取记录数
func (m *Sgorm) GetCount(query interface{}) int64 {
	var count int64
	// 通过条件查询记录数
	if err := m.DB.Model(query).Where(query).Count(&count).Error; err != nil {
		log.Printf("GetCount Error: %v", err)
		return 0
	}
	return count
}

// SelectCount 根据SQL查询记录数
func (m *Sgorm) SelectCount(sql string, values ...interface{}) int64 {
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
func (m *Sgorm) Insert(obj interface{}) error {
	// 创建记录
	return m.DB.Create(obj).Error
}

// InsertBatch 批量插入数据
func (m *Sgorm) InsertBatch(obj []interface{}) error {
	// 批量创建记录
	return m.DB.CreateInBatches(obj, 300).Error
}

// Tx 使用事务执行操作
func (m *Sgorm) Tx(fc func(tx *gorm.DB) error) error {
	return m.DB.Transaction(fc)
}

// SaveOrUpdate 保存或更新记录
func (m *Sgorm) SaveOrUpdate(obj interface{}) *gorm.DB {
	return m.DB.Save(obj)
}

// Update 更新记录
func (m *Sgorm) Update(obj interface{}, query string, args ...interface{}) *gorm.DB {
	return m.DB.Where(query, args...).Updates(obj)
}

// Delete 删除记录，使用ID或其他条件进行删除
func (m *Sgorm) DeleteByField(obj interface{}, query string, args ...interface{}) *gorm.DB {
	return m.DB.Model(obj).Where(query, args...).Delete(obj)
}

// Delete 删除记录，使用ID或其他条件进行删除
func (m *Sgorm) DeleteByObject(obj interface{}) *gorm.DB {
	return m.DB.Model(obj).Where(obj).Delete(obj)
}

// SelectList 查询列表
func (m *Sgorm) SelectList(data interface{}) error {
	return m.DB.Where(data).Find(data).Error
}

// SelectAll 查询所有记录
func (m *Sgorm) SelectAll(data interface{}) error {
	return m.DB.Find(data).Error
}

// SelectOne 查询单条记录
func (m *Sgorm) SelectOne(data interface{}) error {
	return m.DB.Where(data).First(data).Error
}

// SelectSQL 执行SQL查询
func (m *Sgorm) SelectSQL(data interface{}, sql string, values ...interface{}) error {
	return m.DB.Raw(sql, values...).Scan(data).Error
}

// SelectPage 分页查询
func (m *Sgorm) SelectPage(data interface{}, page *model.PageInfo, sql string, values ...interface{}) *model.PageResult {
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

// DeleteById 根据ID删除记录
func DeleteById[T any](id string) error {
	// 根据ID删除记录
	return DB.Model(new(T)).Delete(new(T), id).Error
}

// DeleteByIds 根据多个ID删除记录
func DeleteByIds[T any](ids []interface{}) error {
	// 根据ID数组批量删除记录
	return DB.Model(new(T)).Delete(new(T), ids).Error
}

// SelectList 根据条件查询记录列表
func SelectList[T any](query interface{}) []T {
	var result []T
	// 查询符合条件的记录
	if err := DB.Where(query).Find(&result).Error; err != nil {
		log.Printf("SelectList Error: %v", err)
		return nil
	}
	return result
}

// SelectAll 查询所有记录
func SelectAll[T any]() []T {
	var result []T
	// 查询所有记录
	if err := DB.Find(&result).Error; err != nil {
		log.Printf("SelectAll Error: %v", err)
		return nil
	}
	return result
}

// SelectOne 根据条件查询单条记录
func SelectOne[T any](data interface{}) *T {
	var result T
	// 查询单条记录
	if err := DB.Where(data).First(&result).Error; err != nil {
		log.Printf("SelectOne Error: %v", err)
		return nil
	}
	return &result
}

// Get 根据条件查询记录，返回单个对象
func Get[T any](data interface{}) *T {
	var result T
	// 查询单条记录
	if err := DB.Where(data).First(&result).Error; err != nil {
		log.Printf("Get Error: %v", err)
		return nil
	}
	return &result
}

type GTime struct {
	time.Time
}

// MarshalJSON 自定义时间序列化方法
func (ct GTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format("2006-01-02 15:04:05"))), nil
}

// UnmarshalJSON 自定义时间反序列化方法
func (ct *GTime) UnmarshalJSON(b []byte) error {
	parsedTime, err := time.Parse("2006-01-02 15:04:05", string(b))
	if err != nil {
		return err
	}
	ct.Time = parsedTime
	return nil
}
