package subscriber

import (
	pbgtw "StealthIMMSAP/StealthIM.DBGateway" // 导入 DBGateway protobuf 包
	pb "StealthIMMSAP/StealthIM.MSAP"
	"StealthIMMSAP/config"
	"StealthIMMSAP/gateway"
	stealthimnats "StealthIMMSAP/nats" // 导入 StealthIMMSAP/nats 包并使用别名 stealthimnats
	"fmt"
	"log" // 导入 log 包用于错误处理

	gonats "github.com/nats-io/nats.go" // 导入 nats.go 包并使用别名 gonats
)

// Init 初始化订阅服务
func Init(cfg config.Config) {
	err := stealthimnats.Subscribe(stealthimnats.SubjectMessage, SendMsg)
	if err != nil {
		log.Fatalf("[SUBSCRIBER]Failed to subscribe to NATS message subject: %v", err)
	}

	// 订阅缓存失效通知
	err = stealthimnats.Subscribe(stealthimnats.SubjectCacheInvalidate, InvalidateCache)
	if err != nil {
		log.Fatalf("[SUBSCRIBER]Failed to subscribe to NATS cache invalidate subject: %v", err)
	}
}

// InvalidateCache 处理缓存失效通知
func InvalidateCache(in *pb.N_CacheInvalidateMessage, msg *gonats.Msg) error {
	if in.Groupid == 0 {
		return fmt.Errorf("groupid is empty for CacheInvalidate message")
	}

	// 缓存键应为 msap:msg:latest_id:{groupid}，因为 SyncMessage 缓存的是历史消息，
	// 撤回消息后，最新的消息ID可能不变，但历史消息内容改变，
	// 最直接的失效方式是删除最新消息ID缓存，促使客户端重新获取最新消息ID，
	// 从而可能触发新的历史消息查询。
	// 理想情况下，应该删除所有 msap:msg:history:{groupid}:* 键，但这需要 Redis 的 KEYS 或 SCAN 命令，
	// 而 gateway.ExecRedisDel 不支持模式匹配。
	cacheKey := fmt.Sprintf("msap:msg:latest_id:%d", in.Groupid)

	// 删除 Redis 缓存
	_, err := gateway.ExecRedisDel(&pbgtw.RedisDelRequest{
		DBID: 0, // 假设使用默认DBID
		Key:  cacheKey,
	})
	if err != nil {
		log.Printf("[SUBSCRIBER]Failed to delete latest message ID cache for group %d: %v", in.Groupid, err)
		return err
	}

	// log.Printf("[SUBSCRIBER]Successfully invalidated cache for group %d (msgid: %d)", in.Groupid, in.Msgid) // 取消这类验证日志
	msg.Ack() // 确认消息
	return nil
}
