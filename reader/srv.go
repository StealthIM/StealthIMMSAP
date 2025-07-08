package reader

import (
	pb "StealthIMMSAP/StealthIM.MSAP"
	"StealthIMMSAP/config"
	"StealthIMMSAP/errorcode"
	stealthimnats "StealthIMMSAP/nats" // 使用别名 stealthimnats
	"context"
	"log"
	"time"
)

// WriterKey 写入者标识
var WriterKey = "writer"

// SendMessage 处理发送消息请求
func (*server) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	if config.LatestConfig.SendGRPCProxy.Log {
		log.Printf("[MSAP]Call SendMessage")
	}
	if req.Uid == 0 {
		return &pb.SendMessageResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPUIDEmpty,
				Msg:  "UID is empty",
			},
		}, nil
	}
	if req.Groupid == 0 {
		return &pb.SendMessageResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPGroupIDEmpty,
				Msg:  "GroupID is empty",
			},
		}, nil
	}
	if req.Msg == "" {
		return &pb.SendMessageResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPContentEmpty,
				Msg:  "Content is empty",
			},
		}, nil
	}
	if len(req.Msg) > config.LatestConfig.Database.MaxMsgSize*1024 {
		return &pb.SendMessageResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPContentTooLong,
				Msg:  "Content too long",
			},
		}, nil

	}
	publishMsg := &pb.N_BroadcastMessage{
		Writer:    WriterKey,
		Writetime: time.Now().UnixNano(),
		Action:    pb.N_MessageAction_PreWrite,
		Content: &pb.N_MessageContent{
			Uid:     req.Uid,
			Groupid: req.Groupid,
			Content: req.Msg,
			Type:    req.Type,
			Time:    time.Now().UnixNano(),
		},
	}
	var err error
	for i := range 3 { // 最多重试3次
		err = stealthimnats.Publish(stealthimnats.SubjectMessage, publishMsg)
		if err == nil {
			break // 成功则跳出循环
		}
		if config.LatestConfig.SendGRPCProxy.Log {
			log.Printf("[MSAP]Publish message failed, retrying (%d/3): %v", i+1, err)
		}
		time.Sleep(1 * time.Second) // 重试间隔1秒
	}

	if err != nil {
		switch err.Error() {
		case "No available connections":
			return &pb.SendMessageResponse{
				Result: &pb.Result{
					Code: errorcode.ServerInternalNetworkError,
					Msg:  "No available connections with NATS server",
				},
			}, nil
		default:
			return &pb.SendMessageResponse{
				Result: &pb.Result{
					Code: errorcode.ServerInternalComponentError,
					Msg:  "Server internal component error",
				},
			}, nil
		}
	}
	return &pb.SendMessageResponse{
		Result: &pb.Result{
			Code: errorcode.Success,
			Msg:  "",
		},
	}, nil
}

// RecallMessage 处理撤回消息请求
func (*server) RecallMessage(ctx context.Context, req *pb.RecallMessageRequest) (*pb.RecallMessageResponse, error) {
	if config.LatestConfig.SendGRPCProxy.Log {
		log.Printf("[MSAP]Call RecallMessage")
	}
	if req.Uid == 0 {
		return &pb.RecallMessageResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPUIDEmpty,
				Msg:  "UID is empty",
			},
		}, nil
	}
	if req.Groupid == 0 {
		return &pb.RecallMessageResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPGroupIDEmpty,
				Msg:  "GroupID is empty",
			},
		}, nil
	}
	if req.Msgid == 0 {
		return &pb.RecallMessageResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPMsgIDEmpty,
				Msg:  "MsgID is empty",
			},
		}, nil
	}

	// 发布 NATS 消息通知客户端，由 subscriber 执行数据库更新
	publishMsg := &pb.N_BroadcastMessage{
		Writer:    WriterKey,
		Writetime: time.Now().UnixNano(),
		Action:    pb.N_MessageAction_PreRecall,
		Content: &pb.N_MessageContent{
			Uid:     req.Uid,
			Groupid: req.Groupid,
			Time:    time.Now().UnixNano(),
			Msgid:   req.Msgid,
			Type:    pb.MessageType_Recall_Text, // 使用默认的 Recall_Text 类型
			Content: "",
		},
	}
	var errPublish error
	for i := range 3 { // 最多重试3次
		errPublish = stealthimnats.Publish(stealthimnats.SubjectMessage, publishMsg)
		if errPublish == nil {
			break // 成功则跳出循环
		}
		if config.LatestConfig.SendGRPCProxy.Log {
			log.Printf("[MSAP]Publish recall message failed, retrying (%d/3): %v", i+1, errPublish)
		}
		time.Sleep(1 * time.Second) // 重试间隔1秒
	}

	if errPublish != nil {
		switch errPublish.Error() {
		case "No available connections":
			return &pb.RecallMessageResponse{
				Result: &pb.Result{
					Code: errorcode.ServerInternalNetworkError,
					Msg:  "No available connections with NATS server",
				},
			}, nil
		default:
			return &pb.RecallMessageResponse{
				Result: &pb.Result{
					Code: errorcode.ServerInternalComponentError,
					Msg:  "Server internal component error",
				},
			}, nil
		}
	}
	return &pb.RecallMessageResponse{
		Result: &pb.Result{
			Code: errorcode.Success,
			Msg:  "",
		},
	}, nil
}

// FileAPICall 处理发送消息请求
func (*server) FileAPICall(ctx context.Context, req *pb.FileAPICallRequest) (*pb.FileAPICallResponse, error) {
	if config.LatestConfig.SendGRPCProxy.Log {
		log.Printf("[MSAP]Call SendMessage")
	}
	if req.Uid == 0 {
		return &pb.FileAPICallResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPUIDEmpty,
				Msg:  "UID is empty",
			},
		}, nil
	}
	if req.Groupid == 0 {
		return &pb.FileAPICallResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPGroupIDEmpty,
				Msg:  "GroupID is empty",
			},
		}, nil
	}
	if req.Hash == "" {
		return &pb.FileAPICallResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPContentEmpty,
				Msg:  "Content Hash is empty",
			},
		}, nil
	}
	if req.Filename == "" {
		return &pb.FileAPICallResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPContentEmpty,
				Msg:  "Content Filename is empty",
			},
		}, nil
	}
	publishMsg := &pb.N_BroadcastMessage{
		Writer:    WriterKey,
		Writetime: time.Now().UnixNano(),
		Action:    pb.N_MessageAction_PreWrite,
		Content: &pb.N_MessageContent{
			Uid:      req.Uid,
			Groupid:  req.Groupid,
			Content:  req.Filename,
			Type:     pb.MessageType_File,
			Time:     time.Now().UnixNano(),
			FileHash: req.Hash,
		},
	}
	var err error
	for i := range 3 { // 最多重试3次
		err = stealthimnats.Publish(stealthimnats.SubjectMessage, publishMsg)
		if err == nil {
			break // 成功则跳出循环
		}
		if config.LatestConfig.SendGRPCProxy.Log {
			log.Printf("[MSAP]Publish message failed, retrying (%d/3): %v", i+1, err)
		}
		time.Sleep(1 * time.Second) // 重试间隔1秒
	}

	if err != nil {
		switch err.Error() {
		case "No available connections":
			return &pb.FileAPICallResponse{
				Result: &pb.Result{
					Code: errorcode.ServerInternalNetworkError,
					Msg:  "No available connections with NATS server",
				},
			}, nil
		default:
			return &pb.FileAPICallResponse{
				Result: &pb.Result{
					Code: errorcode.ServerInternalComponentError,
					Msg:  "Server internal component error",
				},
			}, nil
		}
	}
	return &pb.FileAPICallResponse{
		Result: &pb.Result{
			Code: errorcode.Success,
			Msg:  "",
		},
	}, nil
}
