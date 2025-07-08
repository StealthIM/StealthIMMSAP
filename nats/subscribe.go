package nats

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

// Subscribe 订阅 JetStream 消息
func Subscribe[msgType proto.Message](subject Subject, handler func(msgType, *nats.Msg) error) error {
	randomCustomer := fmt.Sprintf("customer-%d", rand.Intn(1000000))
	callback := func(conn *Conn) error {
		subjectName := subjectToString[subject] // 获取主题名称
		// (*conn.JetStream).AddConsumer(subjectName, &nats.ConsumerConfig{
		// 	Durable:       randomCustomer,
		// 	DeliverPolicy: nats.DeliverAllPolicy,
		// 	AckPolicy:     nats.AckExplicitPolicy,
		// 	MaxAckPending: 10,
		// 	AckWait:       30 * time.Second,
		// 	MaxDeliver:    3,
		// })

		// 根据 streams.go 中的 subjectQueue map 判断是否进行队列订阅
		if shouldQueueSubscribe, exists := subjectQueue[subject]; exists && shouldQueueSubscribe {
			// 队列订阅
			queueName := subjectName // 如果需要队列订阅，队列名称与主题名称相同

			reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)
			cleanedQueueName := reg.ReplaceAllString(queueName, "_")
			if cleanedQueueName == "" {
				cleanedQueueName = "default_consumer"
			}

			var dur nats.SubOpt = nats.ManualAck()
			if shouldDurable, exists := subjectDurable[subject]; exists && shouldDurable {
				dur = nats.Durable(randomCustomer)
			}

			_, err := (*conn.JetStream).QueueSubscribe(subjectName, cleanedQueueName+"-queue", func(msg *nats.Msg) {
				// 使用模板创建正确类型的实例 (msgType 预期为指针类型，例如 *MyMessage)
				msg.Ack()
				var zero msgType
				pbMsg := zero.ProtoReflect().New().Interface().(msgType)
				// 将数据反序列化到具体类型的实例中
				if err := proto.Unmarshal(msg.Data, pbMsg); err != nil {
					// 反序列化失败，忽略消息
					return
				}
				// 将具体类型的实例传递给处理函数
				handler(pbMsg, msg)
			},
				nats.ManualAck(),
				nats.MaxAckPending(10),
				nats.AckWait(30*time.Second),
				nats.MaxDeliver(3),
				dur, // 使用 durable 模式
			)
			return err
		}
		// 非队列订阅
		_, err := (*conn.JetStream).Subscribe(subjectName, func(msg *nats.Msg) {
			// 使用模板创建正确类型的实例 (msgType 预期为指针类型，例如 *MyMessage)
			msg.Ack()
			var zero msgType
			pbMsg := zero.ProtoReflect().New().Interface().(msgType)
			// 将数据反序列化到具体类型的实例中
			if err := proto.Unmarshal(msg.Data, pbMsg); err != nil {
				// 反序列化失败，忽略消息
				return
			}
			// 将具体类型的实例传递给处理函数
			go handler(pbMsg, msg)
		})
		return err
	}
	waitSubscribers = append(waitSubscribers, &callback)
	return nil
}

// SubscribeCore 订阅 NATS Core 消息
func SubscribeCore[msgType proto.Message](subject Subject, handler func(msgType)) error {
	callback := func(conn *Conn) error {
		subjectName := subjectToString[subject] // 获取主题名称

		if shouldQueueSubscribe, exists := subjectQueue[subject]; exists && shouldQueueSubscribe {
			// 队列订阅
			queueName := subjectName // 如果需要队列订阅，队列名称与主题名称相同

			// Clean the queueName to ensure it's a valid consumer name
			reg := regexp.MustCompile(`[^a-zA-Z0-9\-_.]+`)
			cleanedQueueName := reg.ReplaceAllString(queueName, "_")
			if cleanedQueueName == "" {
				cleanedQueueName = "default_consumer" // Fallback if cleaning results in empty string
			}

			_, err := conn.Conn.QueueSubscribe(subjectName, cleanedQueueName, func(msg *nats.Msg) {
				// 使用模板创建正确类型的实例 (msgType 预期为指针类型，例如 *MyMessage)
				var zero msgType
				msg.Ack()
				pbMsg := zero.ProtoReflect().New().Interface().(msgType)
				// 将数据反序列化到具体类型的实例中
				if err := proto.Unmarshal(msg.Data, pbMsg); err != nil {
					// 反序列化失败，忽略消息
					return
				}
				// 将具体类型的实例传递给处理函数
				go handler(pbMsg)
			})
			return err
		}
		// 非队列订阅
		_, err := conn.Conn.Subscribe(subjectName, func(msg *nats.Msg) {
			// 使用模板创建正确类型的实例 (msgType 预期为指针类型，例如 *MyMessage)
			var zero msgType
			msg.Ack()
			pbMsg := zero.ProtoReflect().New().Interface().(msgType)
			// 将数据反序列化到具体类型的实例中
			if err := proto.Unmarshal(msg.Data, pbMsg); err != nil {
				// 反序列化失败，忽略消息
				return
			}
			// 将具体类型的实例传递给处理函数
			go handler(pbMsg)
		})
		return err
	}
	waitSubscribers = append(waitSubscribers, &callback)
	return nil
}
