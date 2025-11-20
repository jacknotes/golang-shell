package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

func main() {
	// 配置Kafka消费者
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// 创建消费者实例
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalf("Error closing consumer: %v", err)
		}
	}()

	// 订阅主题
	topic := "test-topic"
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to start consumer for partition 0: %v", err)
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalf("Error closing consumer for partition 0: %v", err)
		}
	}()

	// 监听系统信号，优雅退出
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// 消费消息
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Printf("Consumed message offset %d\n", msg.Offset)
			fmt.Println(string(msg.Value))
		case err := <-partitionConsumer.Errors():
			log.Printf("Error from consumer: %v", err.Err)
		case <-signals:
			return
		}
	}
}
