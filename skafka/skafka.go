package skafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/androidsr/sc-go/syaml"
)

var (
	producer *Producer
)

type Producer struct {
	sarama.SyncProducer
	config   *syaml.KafkaInfo
	Settings sarama.Config
}

// 创建一个消息生产者对象;创建后可设置自定义参数在进行连接操作。
func NewProducer(cfg *syaml.KafkaInfo) *Producer {
	settings := sarama.NewConfig()
	settings.Producer.RequiredAcks = sarama.RequiredAcks(cfg.Producer.RequiredAcks)
	if cfg.Producer.Partitioner == -1 {
		settings.Producer.Partitioner = sarama.NewRandomPartitioner
	} /* else {
		settings.Producer.Partitioner = cfg.Producer.Partitioner
	} */
	settings.Producer.Return.Successes = cfg.Producer.Successes
	settings.Producer.Return.Errors = cfg.Producer.Errors
	settings.Producer.Retry.Max = cfg.Producer.RetryMax
	settings.Producer.Retry.Backoff = time.Duration(cfg.Producer.RetryBackoff) * time.Millisecond
	producer = &Producer{nil, cfg, *settings}
	return producer
}

// 连接kafka服务器
func (m *Producer) Connect() error {
	p, err := sarama.NewSyncProducer(m.config.Nodes, &m.Settings)
	m.SyncProducer = p
	return err
}

func GetProducer() sarama.SyncProducer {
	return producer
}

// 发送消息到指定的主题中,响应分片,移位量,错误信息
func (m Producer) Send(topic string, msg string) (int32, int64, error) {
	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msg),
	}
	partition, offset, err := producer.SendMessage(message)
	return partition, offset, err
}

// 发送消息到指定的主题中,响应分片,移位量,错误信息
func (m Producer) SendOrder(topic string, partition int32, msg string) (int32, int64, error) {
	message := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(msg),
		Partition: partition,
	}
	partition, offset, err := producer.SendMessage(message)
	return partition, offset, err
}

// 关闭生产者
func Close() error {
	return producer.Close()
}

type ConsumerCallback func(message *sarama.ConsumerMessage) bool
type Consumer struct {
	sarama.ConsumerGroup
	config      *syaml.KafkaInfo
	Settings    *sarama.Config
	ctx         context.Context
	groupId     string
	callbackMap map[string]ConsumerCallback
}

func NewConsumer(config *syaml.KafkaInfo) *Consumer {
	settings := sarama.NewConfig()
	settings.Consumer.Offsets.Initial = sarama.OffsetNewest
	settings.Net.MaxOpenRequests = config.Consumer.MaxOpenRequests
	settings.Consumer.Return.Errors = config.Consumer.ReturnErrors
	settings.Consumer.Offsets.AutoCommit.Enable = config.Consumer.AutoCommitEnable
	settings.Consumer.Offsets.AutoCommit.Interval = time.Duration(config.Consumer.AutoCommitInterval) * time.Second // 间隔
	settings.Consumer.Offsets.Retry.Max = config.Consumer.RetryMax
	consumer := new(Consumer)
	consumer.config = config
	consumer.Settings = settings
	consumer.callbackMap = make(map[string]ConsumerCallback, 0)
	return consumer
}

func (m *Consumer) AddBack(topic string, callback ConsumerCallback) {
	m.callbackMap[topic] = callback
}

func (m *Consumer) Listener(ctx context.Context, groupId string, topic []string) error {
	for _, v := range topic {
		if m.callbackMap[v] == nil {
			return fmt.Errorf("主题【%s】未设置消息处理回调", v)
		}
	}
	consumer, err := sarama.NewConsumerGroup(m.config.Nodes, groupId, m.Settings)
	if err != nil {
		return err
	}
	m.ConsumerGroup = consumer
	m.groupId = groupId
	m.ctx = ctx

	go func() {
		for err := range m.Errors() {
			log.Printf("kafka %s consume err:%v", groupId, err)
		}
	}()
	go func() {
		<-ctx.Done()
		m.Close()
	}()
	go func() {
		defer m.Close()
		for {
			fmt.Println("....监听中...")
			err := m.Consume(m.ctx, topic, m)
			if err != nil {
				switch err {
				case sarama.ErrClosedClient, sarama.ErrClosedConsumerGroup:
					log.Printf("quit: kafka consumer %s", m.groupId)
					return
				case sarama.ErrOutOfBrokers:
					log.Fatal("kafka 崩溃了~")
				default:
					log.Printf("kafka exception: %s", err.Error())
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	return nil
}

// Setup 执行在 获得新 session 后 的第一步, 在 ConsumeClaim() 之前
func (c *Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	//fmt.Println("Setup...获得新 session")
	return nil
}

// Cleanup 执行在 session 结束前, 当所有 ConsumeClaim goroutines 都退出时
func (c *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	//fmt.Println("Setup...执行在 session 结束前")
	return nil
}

// ConsumeClaim 具体的消费逻辑
func (c *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	//fmt.Println("具体的消费逻辑")
	for msg := range claim.Messages() {
		result := c.callbackMap[msg.Topic](msg)
		if result {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}
