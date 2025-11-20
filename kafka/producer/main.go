package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

func main() {
	// 配置Kafka生产者
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	// 创建生产者实例
	producer, err := sarama.NewSyncProducer([]string{"172.168.2.17:9192","172.168.2.18:9192","172.168.2.19:9192"}, config)
	if err != nil {
		log.Fatalf("Failed to start producer: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalf("Error closing producer: %v", err)
		}
	}()

	// 监听系统信号，优雅退出
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// 发送消息
	topic := "test-topic"
	for {
		select {
		case <-signals:
			return
		default:
			msg := &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.StringEncoder("Hello, Kafka!"),
			}
			partition, offset, err := producer.SendMessage(msg)
			if err != nil {
				log.Printf("Failed to send message: %v", err)
			} else {
				fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)
			}
			time.Sleep(2 * time.Second)
		}
	}
}
