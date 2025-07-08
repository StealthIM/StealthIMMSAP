package subscriber

import (
	"StealthIMMSAP/config"
	"StealthIMMSAP/nats"
	"log" // 导入 log 包用于错误处理
)

// Init 初始化订阅服务
func Init(cfg config.Config) {
	err := nats.Subscribe(nats.SubjectMessage, SendMsg)
	if err != nil {
		log.Fatalf("[SUBSCRIBER]Failed to subscribe to NATS message subject: %v", err)
	}
}
