## sc-go 框架介绍

目前go生成中还没有一套类似spring boot 的框架；一些功能进行简单的使用也都不是太方便，因此通过个人经验，个人对go中常用的库进行统一整合，并进行封闭扩展，所有组件也都保留了它原有的使用方式，方便个人日常开发学习交流使用。
已适合组件：
gin，sqlx，nacos，go-resty，redis，kafka，minio，yaml，snowflake，elasticsearch

### 安装

```yaml
go get github.com/androidsr/sc-go
```

### yaml集成

框架整个配置文件采用yaml格式进行配置，参考spring boot 将内置参数组件参数进行统一规划整理；

#### 加载配置

```go
// 加载本地配置文件
configs, err := syaml.LoadFile[syaml.PaasRoot]("paas.yaml")
// 自定义数据加载文件，如：nacos配置获取到的数据
configs, err := syaml.Load[syaml.PaasRoot]([]byte(""))
```

### gin集成

通过对gin的扩展封闭，方便日常开发,并扩展路由注册注解实现。

#### 初始化

注册需要将controller包导入进主函数，避免不执行init 方法。也可手动创建controller

```go
import (
    "fmt"

    _ "github.com/androidsr/paas-go/controller"
    "github.com/androidsr/paas-go/sgin"
    "github.com/androidsr/paas-go/syaml"
    "github.com/gin-gonic/gin"
)

router := sgin.New(configs.Paas.Gin)
//router.Use(func(ctx *gin.Context) {
//    fmt.Println("gin插件")
//})

router.RunServer()
```

#### controller

定义一个结构体，并定义方法，可标准gin方式方法，也可以指定参数对接收数据。
注解可按配置文件定义前缀，通过空格 指定 请求类型 请求路径 响应数据类型。
响应数据类型不是必需的。可通过默认gin方式返回数据，也可通过方法返回数据。
在init方法中add router对象。

```go

import (
    "fmt"

    "github.com/androidsr/paas-go/sgin"
    "github.com/gin-gonic/gin"
)

func init() {
    sgin.AddRouter(IndexController{})
}

type IndexController struct {
}

// @Router [post] / [json]
func (IndexController) Index(c *gin.Context, data *User) (string, error) {
    fmt.Println("index.........")
    fmt.Println(data)
    return "成功", nil
}

type User struct {
    Name string `form:"name" json:"name"`
    Sex  string `form:"sex" json:"sex"`
}

```

### sqlx集成

使用过go中orm各种难受不习惯，相对来说sqlx更适合。但是经常写sql也是一个麻烦的事。因此舍弃一些性能提升部分效率是值得的。

#### 初始化

```go
//创建一个连接对象并指定给默认DB对象方便代码中使用。多数据源时不指定，自由管理。
sorm.DB = sorm.New(configs.Paas.Sqlx)

//自动填充配置（参考mybatis-plus）对非空字段进行自动填充
AddAutoFill("id", func() any {
    return snowflake.GetString()
})

```

#### 模型定义

基础模型定义参考sqlx标准功能，扩展keyword tag；用做查询条件组装。参考mybatis-plus queryWrapper

```go
type SysButtons struct {
    Id      string `json:"id" db:"id,primary_key" keyword:"eq,id"`
    Title   string `json:"title" keyword:"like"`
    Click   string `json:"click"`
    Icon    string `json:"icon"`
    State   string `json:"state"`
    OrderId string `json:"orderId" db:"order_id"`
}

```

#### 常规操作

```go
//插入数据
data := new(SysButtons)
data.Title = "测试"
data.Click = "事件"
data.State = "1"
data.OrderId = "1"
fmt.Println(DB.Insert(data))

//更新数据
query := &SysButtons{Id: "1656565833582776320"}
var data = new(SysButtons)
DB.SelectOne(data, query)
data.Title = "测试1"
data.Click = "事件1"
data.State = "1"
data.OrderId = "12"
//参数说明：更新对象,条件字段列
fmt.Println(DB.Update(data, "id"))

//更新数据
query := &SysButtons{Id: "1656565833582776320"}
var data = new(SysButtons)
DB.SelectOne(data, query)
data.Title = "测试1"
data.Click = "事件1"
data.State = "1"
data.OrderId = "12"
//参数说明：更新对象,条件字段列
fmt.Println(DB.UpdateById(data))

//删除数据（因为对象是必需的，条件是必需的因此就不构建byId方法了）
data := new(SysButtons)
data.Id = "1656533792241750016"
fmt.Println(DB.Delete(data))

//分页查询（配合keyword进行条件查询，通过别名可进行表连接条件）
query := new(SysButtons)
query.State = "1"
//query.Title = "增"
sql := `select * from sys_buttons a where 1=1 `
var data []SysButtons
v := DB.SelectPage(&data, sql, query, &model.PageInfo{Current: 1, Size: 10})
fmt.Println(v)//返回已包装好的page对象
fmt.Println(data)//返回纯数据对象
```

### nacos集成

集成nacos配置中心和注册中心方便服务调用。

```go
//初始化并注册服务
New(configs.Paas.Nacos)
```

##### 配置中心（仅支持nacos 2.0及以上）

```go
//获取一个配置文件回调处理。
ConfigClient.GetDefaultConfig(func(namespace, group, dataId, data string) {
    fmt.Println(namespace, group, dataId, data)
})
```

##### 注册中心

```go
//获取一个可用的服务
instance, err := NamingClient.GetInstance("paas-go")
if err != nil {
    panic(err)
}
fmt.Println(instance.Ip, instance.Port)

//通过服务名获取url地址（可通过hcli或http工具进行简单的服务调用）
url := nacos.GetUrl("服务名","/接口路径")
```

### hcli(go-resty)集成

集成resty进行服务调用，并定义默认方法

#### 初始化

```go
//创建客户端并指定header,cookie,可不指定。
cli:=hcli.New(nil,nil)

```

##### 接口调用

```go
var dest interface{}//json格式各应对象，也可不指定
url:=nacos.GetUrl("","")
//url:= "http://www.baidu.com"
bs,err:=cli.Get(dest,url,params)
//响应结果支持指定参数，也可通过方法返回的结果[]byte数据自行处理。
//其它操作类似
```

### kafka集成

#### 生产者

生产者保持全局唯一，理论上不需要创建多个实例。

```go
configs, _ := syaml.LoadFile[syaml.PaasRoot]("../paas.yaml")
producer := NewProducer(configs.Paas.Kafka)
err := producer.Connect()
if err != nil {
    panic(err)
}
for i := 0; i < 5; i++ {
    i, v, err := producer.Send(fmt.Sprintf("%s%d", "test", i), "你好kafka")
    fmt.Println(i, v, err)
}
```

#### 消费者

消费端以group为单位，不同分组下创建不同的tcp连接，通过统一监听主题，回调处理的方式进行业务处理

```go
configs, err := syaml.LoadFile[syaml.PaasRoot]("../paas.yaml")
if err != nil {
    panic(err)
}
consumer := NewConsumer(configs.Paas.Kafka)
consumer.Settings.Consumer.Offsets.Initial = sarama.OffsetOldest
consumer.AddBack("test0", func(message *sarama.ConsumerMessage) bool {
    time.Sleep(5 * time.Second)
    fmt.Println("消息处理：", string(message.Key), message.Topic, message.Partition, message.Offset, string(message.Value), paas.FormatDateTimeString(time.Now()))
    return true
})
consumer.AddBack("test1", func(message *sarama.ConsumerMessage) bool {
    fmt.Println("消息处理：", string(message.Key), message.Topic, message.Partition, message.Offset, string(message.Value), paas.FormatDateTimeString(time.Now()))
    return true
})
consumer.AddBack("test2", func(message *sarama.ConsumerMessage) bool {
    fmt.Println("消息处理：", string(message.Key), message.Topic, message.Partition, message.Offset, string(message.Value), paas.FormatDateTimeString(time.Now()))
    return true
})
consumer.AddBack("test3", func(message *sarama.ConsumerMessage) bool {
    fmt.Println("消息处理：", string(message.Key), message.Topic, message.Partition, message.Offset, string(message.Value), paas.FormatDateTimeString(time.Now()))
    return true
})
consumer.AddBack("test4", func(message *sarama.ConsumerMessage) bool {
    fmt.Println("消息处理：", string(message.Key), message.Topic, message.Partition, message.Offset, string(message.Value), paas.FormatDateTimeString(time.Now()))
    return true
})
err = consumer.Listener(context.Background(), "default_group", []string{"test0", "test1", "test2", "test3", "test4"})
if err != nil {
    panic(err)
}
```

### websocket集成

通过与任意http服务进行集成，转websocket服务。由gorilla/websocket实现。内置心跳处理，双工逻辑等

```go
configs, _ := syaml.LoadFile[syaml.PaasRoot]("../paas.yaml")
router := sgin.New(configs.Paas.Gin)
socket := New(websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}, time.Second*60, 3, func(w http.ResponseWriter, r *http.Request) string {
    return "sirui"
})
router.GET("/ws", func(c *gin.Context) {
    err := socket.Register(c.Writer, c.Request)
    if err != nil {
        c.JSON(http.StatusOK, model.NewFail(5000, err.Error()))
    }
})
//接收消息
go func() {
    for {
        m := <-socket.Data
        fmt.Println(m.UserId, string(m.Data))
        socket.Write(m.UserId, websocket.TextMessage, []byte("你好，"+string(m.Data)))
    }
}()

```

### redis集成

### minio集成

待开发...
