package subscriber

import (
	pbgtw "StealthIMMSAP/StealthIM.DBGateway"
	pb "StealthIMMSAP/StealthIM.MSAP"
	"StealthIMMSAP/errorcode"
	"StealthIMMSAP/gateway"
	stealthimnats "StealthIMMSAP/nats" // 使用别名 stealthimnats
	"fmt"
	"log"
	"strconv"
	"time"

	gonats "github.com/nats-io/nats.go" // 使用别名 gonats
)

// SendMsg 写入消息到数据库或处理撤回消息
func SendMsg(in *pb.N_BroadcastMessage, msg *gonats.Msg) error {
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

	switch in.Action {
	case pb.N_MessageAction_PreWrite:
		// 处理普通消息写入
		if in.Content.Content == "" {
			return fmt.Errorf("content is empty for PreWrite message")
		}
		if in.Content.Time < time.Now().Unix()-120 {
			return fmt.Errorf("time is incorrect for PreWrite message")
		}
		// 拒绝 recall 类型的普通消息
		if in.Content.Type >= 16 {
			return fmt.Errorf("recall messages should be sent via RecallMessage request, not SendMessage")
		}

		var msgid int64 = -1

		// 写入消息到数据库
		sql := "INSERT INTO msg (group_id, msg_content, msg_msgTime, msg_uid, msg_fileHash, msg_type, msg_sender) VALUES (?, ?, FROM_UNIXTIME(?), ?, ?, ?, ?)"
		params := []*pbgtw.InterFaceType{
			{Response: &pbgtw.InterFaceType_Int64{Int64: in.Content.Groupid}},
			{Response: &pbgtw.InterFaceType_Str{Str: in.Content.Content}},
			{Response: &pbgtw.InterFaceType_Int64{Int64: in.Content.Time / 1000000000}},
			{Response: &pbgtw.InterFaceType_Int32{Int32: in.Content.Uid}},
			{Response: &pbgtw.InterFaceType_Str{Str: in.Content.FileHash}}, // 使用 N_MessageContent 中的 file_hash
			{Response: &pbgtw.InterFaceType_Int32{Int32: int32(in.Content.Type)}},
			{Response: &pbgtw.InterFaceType_Str{Str: (in.Content.Username)}},
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
		err = stealthimnats.Publish(stealthimnats.SubjectCacheInvalidate, invalidateMsg)
		if err != nil {
			log.Printf("[SUBSCRIBER]Failed to publish cache invalidate message: %v", err)
		}

		// 更新最新消息ID缓存
		latestMsgIDKey := fmt.Sprintf("msap:msg:latest_id:%d", in.Content.Groupid)
		_, err = gateway.ExecRedisSet(&pbgtw.RedisSetStringRequest{
			DBID:  0,
			Key:   latestMsgIDKey,
			Value: strconv.FormatInt(msgid, 10),
			Ttl:   60,
		})
		if err != nil {
			log.Printf("[SUBSCRIBER]Failed to set latest message ID cache for group %d: %v", in.Content.Groupid, err)
		}

		// 广播普通消息给对应组的在线用户
		for i := range 3 { // 最多重试3次
			err = stealthimnats.PublishWithSubject(fmt.Sprintf("push.%d", in.Content.Groupid), &pb.SyncMessageResponse{
				Result: &pb.Result{Code: errorcode.Success, Msg: ""},
				Time:   in.Content.Time,
				Msg: []*pb.ReciveMessageListen{
					{
						Msgid:    msgid,
						Uid:      in.Content.Uid,
						Groupid:  in.Content.Groupid,
						Msg:      in.Content.Content,
						Type:     in.Content.Type,
						Hash:     in.Content.FileHash, // 使用 N_MessageContent 中的 file_hash
						Time:     in.Content.Time / 1000000000,
						Username: in.Content.Username,
					},
				},
			})
			if err == nil {
				break
			}
			log.Printf("[SUBSCRIBER]Publish message failed, retrying (%d/3): %v", i+1, err)
			time.Sleep(1 * time.Second)
		}

		if err != nil {
			log.Printf("[SUBSCRIBER]Failed to publish message after retries: %v", err)
			return fmt.Errorf("failed to publish message")
		}
		return nil

	case pb.N_MessageAction_PreRecall:
		// 处理撤回消息
		if in.Content.Msgid == 0 {
			return fmt.Errorf("msgid is empty for PreRecall message")
		}
		if in.Content.Uid == 0 { // 添加 UID 校验
			return fmt.Errorf("uid is empty for PreRecall message")
		}

		// 执行数据库更新：将消息类型更新为撤回类型
		// UPDATE msg SET msg_type = msg_type + 16 WHERE msg_id = ? AND group_id = ? AND msg_uid = ? AND msg_type < 16
		sqlUpdate := "UPDATE msg SET msg_type = msg_type + 16 WHERE msg_id = ? AND group_id = ? AND msg_uid = ? AND msg_type < 16"
		updateRes, err := gateway.ExecSQL(&pbgtw.SqlRequest{
			Sql: sqlUpdate,
			Db:  pbgtw.SqlDatabases_Msg,
			Params: []*pbgtw.InterFaceType{
				{Response: &pbgtw.InterFaceType_Int64{Int64: in.Content.Msgid}},
				{Response: &pbgtw.InterFaceType_Int64{Int64: in.Content.Groupid}},
				{Response: &pbgtw.InterFaceType_Int32{Int32: in.Content.Uid}},
			},
			Commit:      true,
			GetRowCount: true,
		})

		if err != nil || updateRes.Result.Code != errorcode.Success {
			log.Printf("[SUBSCRIBER]Failed to update message type in database for recall: %v, SQL Result Code: %d, Msg: %s", err, updateRes.Result.Code, updateRes.Result.Msg)
			return fmt.Errorf("failed to update message type in database for recall")
		}

		// 如果没有行受影响，说明消息不存在、UID不匹配或消息已被recall
		if updateRes.RowsAffected == 0 {
			// 这里不返回错误，因为可能是重复的撤回请求或者消息不符合条件，但NATS消息已经处理
			return nil
		}

		// 发布缓存失效通知
		invalidateMsg := &pb.N_CacheInvalidateMessage{
			Groupid: in.Content.Groupid,
			Msgid:   in.Content.Msgid, // 使用被撤回的消息ID
		}
		err = stealthimnats.Publish(stealthimnats.SubjectCacheInvalidate, invalidateMsg)
		if err != nil {
			log.Printf("[SUBSCRIBER]Failed to publish cache invalidate message for recall: %v", err)
		}

		// 广播撤回消息给对应组的在线用户
		for i := range 3 { // 最多重试3次
			err = stealthimnats.PublishWithSubject(fmt.Sprintf("push.%d", in.Content.Groupid), &pb.SyncMessageResponse{
				Result: &pb.Result{Code: errorcode.Success, Msg: ""},
				Time:   in.Content.Time,
				Msg: []*pb.ReciveMessageListen{
					{
						Msgid:    in.Content.Msgid,
						Uid:      in.Content.Uid,
						Groupid:  in.Content.Groupid,
						Type:     in.Content.Type, // 应该是 recall 类型
						Time:     in.Content.Time / 1000000000,
						Username: in.Content.Username,
						// Msg, Hash, Username 字段留空
					},
				},
			})
			if err == nil {
				break
			}
			log.Printf("[SUBSCRIBER]Publish recall message failed, retrying (%d/3): %v", i+1, err)
			time.Sleep(1 * time.Second)
		}

		if err != nil {
			log.Printf("[SUBSCRIBER]Failed to publish recall message after retries: %v", err)
			return fmt.Errorf("failed to publish recall message")
		}
		return nil

	default:
		return fmt.Errorf("unknown message action")
	}
}
