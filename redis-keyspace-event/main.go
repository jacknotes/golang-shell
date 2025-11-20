package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client
var ctx = context.Background()
var redis_db = 2

func main() {
	// 创建 Redis 客户端
	rdb = redis.NewClient(&redis.Options{
		// Addr: "redis1_m.hs.com:6369", // Redis 地址
		Addr: "fat-redis.hs.com:6002", // Redis 地址
		DB:   redis_db,                // 默认数据库
		// Password: "pass123",
	})

	// 检查 Redis 连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("无法连接 Redis: %v", err)
	}

	// 订阅键空间通知事件
	pubsub := rdb.Subscribe(ctx, fmt.Sprintf("__keyevent@%d__:expired", redis_db), fmt.Sprintf("__keyevent@%d__:del", redis_db))

	// 监听事件
	go listenToKeyspaceEvents(pubsub)

	// 等待退出
	select {}
}

// 监听键空间事件
func listenToKeyspaceEvents(pubsub *redis.PubSub) {
	for msg := range pubsub.Channel() {
		currentTime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("时间: %s, 接收到事件: %s -> %s\n", currentTime, msg.Channel, msg.Payload)
		// 在这里处理事件，比如记录删除键的时间、执行其他逻辑等
	}
}
