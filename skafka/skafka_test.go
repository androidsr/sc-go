package skafka

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/androidsr/paas-go/paas"
	"github.com/androidsr/paas-go/syaml"
)

func Test_NewProducer(t *testing.T) {
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
}

func Test_NewConsumer(t *testing.T) {
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
	fmt.Println("....等待结果.....")
	time.Sleep(10 * time.Second)
}
