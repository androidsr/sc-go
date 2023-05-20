package sredis

import (
	"context"
	"time"

	"github.com/androidsr/sc-go/syaml"
	"github.com/redis/go-redis/v9"
)

var (
	config        *syaml.RedisInfo
	defaultClient redis.UniversalClient
	client        *SRedis
)

// 创建连接
func New(cfg *syaml.RedisInfo) {
	if defaultClient == nil {
		config = cfg
		defaultClient = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:        config.Nodes,
			MasterName:   config.Master,
			Password:     config.Password,
			DB:           config.Database,
			PoolSize:     config.Pool.PoolSize,
			MinIdleConns: config.Pool.MinIdleConns,
			MaxIdleConns: config.Pool.MaxIdleConns,
			DialTimeout:  time.Duration(config.Pool.DialTimeout) * time.Microsecond,  // 设置连接超时
			ReadTimeout:  time.Duration(config.Pool.ReadTimeout) * time.Microsecond,  // 设置读取超时
			WriteTimeout: time.Duration(config.Pool.WriteTimeout) * time.Microsecond, // 设置写入超时
		})
		client = &SRedis{defaultClient}
	}
}

// 获取go-redis标准客户端
func GetDefault() redis.UniversalClient {
	return defaultClient
}

// 获取redis包装后的客户端
func GetClient() *SRedis {
	return client
}

type SRedis struct {
	client redis.UniversalClient
}

func (m *SRedis) Set(key string, value interface{}, expiration time.Duration) error {
	return m.client.Set(context.Background(), key, value, 0).Err()
}

func (m *SRedis) Get(key string) (string, error) {
	return m.client.Get(context.Background(), key).Result()
}

func (m *SRedis) GetScan(dest interface{}, key string) error {
	return m.client.Get(context.Background(), key).Scan(dest)
}

func (m *SRedis) Delete(key ...string) error {
	return m.client.Del(context.Background(), key...).Err()
}

func (m *SRedis) Exists(key ...string) (bool, error) {
	exists, err := m.client.Exists(context.Background(), key...).Result()
	return exists == 1, err
}

func (m *SRedis) Expire(key string, expiration time.Duration) error {
	err := m.client.Expire(context.Background(), key, expiration).Err()
	return err
}

func (m *SRedis) Subscribe(key string) *redis.PubSub {
	pubsub := m.client.Subscribe(context.Background(), key)
	return pubsub
}

// 从列表左侧插入一个元素
func (m *SRedis) LPush(key string, values ...interface{}) error {
	err := m.client.LPush(context.Background(), key, values...).Err()
	return err
}

// 从列表左侧获取一个元素
func (m *SRedis) LPop(key string) (string, error) {
	result, err := m.client.LPop(context.Background(), key).Result()
	return result, err
}

// 从列表左侧获取一个元素
func (m *SRedis) LPopScan(dest interface{}, key string) error {
	err := m.client.LPop(context.Background(), key).Scan(dest)
	return err
}

// 从列表右侧插入多个元素
func (m *SRedis) RPush(key string, values ...interface{}) error {
	err := m.client.RPush(context.Background(), key, values...).Err()
	return err
}

// 从列表右侧获取一个元素
func (m *SRedis) RPop(key string) (string, error) {
	return m.client.RPop(context.Background(), key).Result()
}

// 从列表右侧获取一个元素
func (m *SRedis) RPopScan(dest interface{}, key string) error {
	err := m.client.RPop(context.Background(), key).Scan(dest)
	return err
}

// 获取列表长度
func (m *SRedis) LLen(key string) (int64, error) {
	length, err := m.client.LLen(context.Background(), key).Result()
	return length, err
}

// 获取列表指定范围内的元素
func (m *SRedis) LRange(key string, start, end int64) ([]string, error) {
	vals, err := m.client.LRange(context.Background(), key, start, end).Result()
	return vals, err
}

// 获取列表指定范围内的元素
func (m *SRedis) LRangeScan(dest []interface{}, key string, start, end int64) error {
	err := m.client.LRange(context.Background(), key, start, end).ScanSlice(dest)
	return err
}

// 设置哈希表的字段值
func (m *SRedis) HSet(key string, values ...interface{}) error {
	err := m.client.HSet(context.Background(), key, values...).Err()
	return err
}

// 设置哈希表的字段值
func (m *SRedis) HGetAll(key string) (map[string]string, error) {
	vals, err := m.client.HGetAll(context.Background(), key).Result()
	return vals, err
}

// 设置哈希表的字段值
func (m *SRedis) HGetAllScan(dest interface{}, key string) error {
	err := m.client.HGetAll(context.Background(), key).Scan(dest)
	return err
}

// 删除哈希表的一个或多个字段
func (m *SRedis) HDel(key string, fields ...string) error {
	err := m.client.HDel(context.Background(), key, fields...).Err()
	return err
}

// 添加一个元素到有序集合
func (m *SRedis) ZAdd(key string, score float64, value interface{}) error {
	err := m.client.ZAdd(context.Background(), key, redis.Z{Score: score, Member: value}).Err()
	return err
}

// 添加多个元素到有序集合
func (m *SRedis) ZAdds(key string, value ...redis.Z) error {
	err := m.client.ZAdd(context.Background(), key, value...).Err()
	return err
}

// 获取有序集合的所有元素
func (m *SRedis) ZRange(key string, start, end int64) ([]string, error) {
	vals, err := m.client.ZRange(context.Background(), key, start, end).Result()
	return vals, err
}

// 获取有序集合的所有元素
func (m *SRedis) ZRangeScan(dest []interface{}, key string, start, end int64) error {
	err := m.client.ZRange(context.Background(), key, start, end).ScanSlice(dest)
	return err
}

// 删除有序集合中一个或多个元素
func (m *SRedis) ZRem(key string, members ...interface{}) error {
	err := m.client.ZRem(context.Background(), key, members...).Err()
	return err
}
