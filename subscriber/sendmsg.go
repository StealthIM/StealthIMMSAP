package subscriber

import (
	pbgtw "StealthIMMSAP/StealthIM.DBGateway"
	pb "StealthIMMSAP/StealthIM.MSAP"
	"StealthIMMSAP/errorcode"
	"StealthIMMSAP/gateway"
	natsmg "StealthIMMSAP/nats"
	"fmt"
	"log"
	"strconv" // 导入 strconv 包
	"time"

	"github.com/nats-io/nats.go"
)

// SendMsg 写入消息到数据库
func SendMsg(in *pb.N_BroadcastMessage, msg *nats.Msg) error {
	if in.Content == nil {
		return fmt.Errorf("Content is nil")
	}
	if in.Content.Uid == 0 {
		return fmt.Errorf("UID is empty")
	}
	if in.Content.Time == 0 {
		return fmt.Errorf("time is empty")
	}
	if in.Content.Groupid == 0 {
		return fmt.Errorf("groupid is empty")
	}
	if in.Content.Content == "" {
		return fmt.Errorf("content is empty")
	}
	if in.Content.Time < time.Now().Unix()-120 {
		return fmt.Errorf("time is incorrect")
	}

	var msgid int64 = -1

	// 写入消息到数据库
	sql := "INSERT INTO msg (group_id, msg_content, msg_msgTime, msg_uid, msg_fileHash, msg_type) VALUES (?, ?, FROM_UNIXTIME(?), ?, ?, ?)"
	params := []*pbgtw.InterFaceType{
		{Response: &pbgtw.InterFaceType_Int64{Int64: in.Content.Groupid}},
		{Response: &pbgtw.InterFaceType_Str{Str: in.Content.Content}},
		{Response: &pbgtw.InterFaceType_Int64{Int64: in.Content.Time / 1000000000}},
		{Response: &pbgtw.InterFaceType_Int32{Int32: in.Content.Uid}},
		{Response: &pbgtw.InterFaceType_Str{Str: ""}}, // file_hash 始终为空
		{Response: &pbgtw.InterFaceType_Int32{Int32: int32(in.Content.Type)}},
	}

	var sqlRes *pbgtw.SqlResponse
	var err error
	for range 3 { // 最多重试3次
		sqlRes, err = gateway.ExecSQL(&pbgtw.SqlRequest{
			Sql:             sql,
			Db:              pbgtw.SqlDatabases_Msg,
			Params:          params,
			Commit:          true,
			GetLastInsertId: true,
		})
		if err == nil && sqlRes.Result.Code == errorcode.Success {
			msgid = sqlRes.LastInsertId
			break // 成功则跳出循环
		}
		time.Sleep(1 * time.Second) // 重试间隔1秒
	}

	if err != nil || sqlRes.Result.Code != errorcode.Success {
		return fmt.Errorf("failed to write message to database")
	}

	// 消息成功写入数据库后，发布缓存失效通知
	invalidateMsg := &pb.N_CacheInvalidateMessage{
		Groupid: in.Content.Groupid,
		Msgid:   msgid,
	}
	err = natsmg.Publish(natsmg.SubjectCacheInvalidate, invalidateMsg)
	if err != nil {
		log.Printf("[SUBSCRIBER]Failed to publish cache invalidate message: %v", err)
		// 即使缓存失效通知失败，也不应该阻止消息推送，因为缓存是优化，不是核心功能
	}

	// 更新最新消息ID缓存
	latestMsgIDKey := fmt.Sprintf("msap:msg:latest_id:%d", in.Content.Groupid)
	_, err = gateway.ExecRedisSet(&pbgtw.RedisSetStringRequest{
		DBID:  0, // 假设 DBID 为 0
		Key:   latestMsgIDKey,
		Value: strconv.FormatInt(msgid, 10),
		Ttl:   60, // 设置较短的 TTL，例如 60 秒，或者每次更新时刷新
	})
	if err != nil {
		log.Printf("[SUBSCRIBER]Failed to set latest message ID cache for group %d: %v", in.Content.Groupid, err)
	}

	for i := range 3 { // 最多重试3次
		err = natsmg.PublishWithSubject(fmt.Sprintf("push.%d", in.Content.Groupid), &pb.SyncMessageResponse{
			Result: &pb.Result{Code: errorcode.Success, Msg: ""},
			Time:   in.Content.Time,
			Msg: []*pb.ReciveMessageListen{
				{
					Msgid:   msgid,
					Uid:     in.Content.Uid,
					Groupid: in.Content.Groupid,
					Msg:     in.Content.Content,
					Type:    in.Content.Type,
					Hash:    "", // 文件哈希始终为空
					Time:    in.Content.Time / 1000000000,
				},
			},
		})
		if err == nil {
			break // 成功则跳出循环
		}
		log.Printf("[SUBSCRIBER]Publish message failed, retrying (%d/3): %v", i+1, err)
		time.Sleep(1 * time.Second) // 重试间隔1秒
	}

	if err != nil {
		log.Printf("[SUBSCRIBER]Failed to publish message after retries: %v", err)
		return fmt.Errorf("failed to publish message")
	}

	return nil
}
