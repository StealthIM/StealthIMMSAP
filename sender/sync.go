package sender

import (
	pbgtw "StealthIMMSAP/StealthIM.DBGateway"
	pb "StealthIMMSAP/StealthIM.MSAP"
	"StealthIMMSAP/config"
	"StealthIMMSAP/errorcode"
	"StealthIMMSAP/gateway"
	natsmg "StealthIMMSAP/nats"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto" // 导入 proto 包
)

const int64Max = 9223372036854775806

// MaxConnections 每个组的最大连接数
const MaxConnections = 1000

// CacheTTL 缓存过期时间（秒）
const CacheTTL = 300 // 5分钟

// GetMessageOnce 单次获取的消息条数
const GetMessageOnce = 32

// MaxLimit 最大获取消息条数
const MaxLimit = 256

// ClientConnection 表示单个客户端的 gRPC 流连接
type ClientConnection struct {
	stream   grpc.ServerStreamingServer[pb.SyncMessageResponse]
	sendChan chan *pb.SyncMessageResponse // sendChan 用于向此客户端发送消息的通道
	done     chan struct{}                // done 用于指示连接完成的通道
}

// GroupConnections 管理特定组的所有客户端连接
type GroupConnections struct {
	mu          sync.Mutex
	connections []*ClientConnection
}

// OnlineUsers 存储所有活跃的组连接
// 键: groupid (int64), 值: *GroupConnections
var OnlineUsers sync.Map

// SyncMessage 处理消息同步的 gRPC 流 管理客户端连接并将消息分派到相应的组
func (s *server) SyncMessage(req *pb.SyncMessageRequest, stream grpc.ServerStreamingServer[pb.SyncMessageResponse]) error {
	if config.LatestConfig.SyncGRPCProxy.Log {
		log.Printf("[SYNCER]Call SyncMessage")
	}

	if req.Groupid == 0 {
		stream.Send(&pb.SyncMessageResponse{
			Result: &pb.Result{
				Code: errorcode.MSAPGroupIDEmpty,
				Msg:  "GroupID is empty",
			}})
		time.Sleep(time.Millisecond * 500)
		return nil
	}

	if req.Sync {
		// 加载或为该 groupid 创建 GroupConnections
		groupValue, _ := OnlineUsers.LoadOrStore(req.Groupid, &GroupConnections{})
		groupConns := groupValue.(*GroupConnections)

		groupConns.mu.Lock()
		if len(groupConns.connections) >= MaxConnections {
			groupConns.mu.Unlock()
			if config.LatestConfig.SyncGRPCProxy.Log {
				log.Printf("[SYNCER]Group %d reached max connections %d", req.Groupid, MaxConnections)
			}
			stream.Send(&pb.SyncMessageResponse{
				Result: &pb.Result{
					Code: errorcode.ServerOverload,
					Msg:  "Group connection limit reached",
				}})
			time.Sleep(time.Millisecond * 500)
			return nil
		}

		// 创建一个新的客户端连接
		clientConn := &ClientConnection{
			stream:   stream,
			sendChan: make(chan *pb.SyncMessageResponse, 100), // 用于消息的缓冲通道
			done:     make(chan struct{}),
		}

		// 将新的客户端连接添加到组中
		groupConns.connections = append(groupConns.connections, clientConn)
		groupConns.mu.Unlock()

		// 启动一个 goroutine 向此客户端发送消息
		go func() {
			for {
				select {
				case msg := <-clientConn.sendChan:
					if err := clientConn.stream.Send(msg); err != nil {
						if config.LatestConfig.SyncGRPCProxy.Log {
							log.Printf("[SYNCER]Error sending message to client: %v", err)
						}
						return
					}
				case <-clientConn.done:
					// 连接已完成 退出 goroutine
					return
				}
			}
		}()

		// 客户端断开连接 进行清理
		defer func() {
			groupConns.mu.Lock()
			for i, conn := range groupConns.connections {
				if conn == clientConn {
					// 移除已断开连接的客户端连接
					groupConns.connections = append(groupConns.connections[:i], groupConns.connections[i+1:]...)
					close(clientConn.done)
					close(clientConn.sendChan)
					break
				}
			}
			groupConns.mu.Unlock()
		}()
	}
	if req.Limit > MaxLimit {
		req.Limit = MaxLimit
	}

	var lastMessageID int64 = req.LastMsgid

	if req.LastMsgid == 0 {
		lastMessageID = int64Max
	}

	if req.LastMsgid < -1 {
		lastMessageID = int64Max
	}

	receivedMsgNum := 0

	if req.LastMsgid == -1 {
		// 无需加载历史
	} else {
		for {
			// 构建 Redis 缓存键
			cacheKey := fmt.Sprintf("msap:msg:history:%d:%d", req.Groupid, lastMessageID)

			// 尝试从 Redis 缓存中获取消息
			getRes, err := gateway.ExecRedisBGet(&pbgtw.RedisGetBytesRequest{
				DBID: 0, // 假设 DBID 为 0
				Key:  cacheKey,
			})

			var messagesToSend []*pb.ReciveMessageListen
			var fromCache bool = false

			if err == nil && getRes.Result.Code == errorcode.Success && len(getRes.Value) > 0 {
				// 缓存命中
				cachedList := &pb.CachedMessageList{}
				if err := proto.Unmarshal(getRes.Value, cachedList); err == nil {
					messagesToSend = cachedList.Messages
					if len(messagesToSend) > 0 {
						fromCache = true
					}
				} else {
					// 反序列化失败，继续从数据库查询
				}
			}

			if !fromCache {
				// 缓存未命中或反序列化失败，从数据库查询
				var sqlx string
				sqlx = fmt.Sprintf("SELECT msg_id, group_id, msg_content, msg_msgTime, msg_uid, msg_fileHash, msg_type, msg_sender FROM msg WHERE group_id = ? AND msg_id < ? ORDER BY msg_id DESC LIMIT %d;", GetMessageOnce)
				sqlReq := &pbgtw.SqlRequest{
					Sql: sqlx, // 每次最多拉取32条
					Db:  pbgtw.SqlDatabases_Msg,
					Params: []*pbgtw.InterFaceType{
						{Response: &pbgtw.InterFaceType_Int64{Int64: req.Groupid}},
						{Response: &pbgtw.InterFaceType_Int64{Int64: int64(lastMessageID)}},
					},
				}

				sqlRes, err := gateway.ExecSQL(sqlReq)
				if err != nil {
					stream.Send(&pb.SyncMessageResponse{
						Result: &pb.Result{
							Code: errorcode.ServerInternalComponentError,
							Msg:  "Failed to query historical messages",
						}})
					return err
				}

				if sqlRes.Result.Code != errorcode.Success {
					stream.Send(&pb.SyncMessageResponse{
						Result: &pb.Result{
							Code: sqlRes.Result.Code,
							Msg:  sqlRes.Result.Msg,
						},
						Time: time.Now().Unix(),
					})
					return fmt.Errorf("failed to query historical messages: %s", sqlRes.Result.Msg)
				}

				if len(sqlRes.Data) == 0 {
					// 没有更多消息，退出循环
					// 缓存空结果，避免缓存穿透
					emptyCachedList := &pb.CachedMessageList{}
					emptyBytes, _ := proto.Marshal(emptyCachedList)
					gateway.ExecRedisBSet(&pbgtw.RedisSetBytesRequest{
						DBID:  0,
						Key:   cacheKey,
						Value: emptyBytes,
						Ttl:   CacheTTL, // 缓存空结果也设置 TTL
					})
					break
				}

				messagesToSend = make([]*pb.ReciveMessageListen, 0, len(sqlRes.Data))
				for _, row := range sqlRes.Data {
					// receivedMsgNum++
					// if receivedMsgNum > int(req.Limit) {
					// 	break
					// }
					msgID := row.Result[0].GetUint64()
					groupID := row.Result[1].GetUint32()
					msgContent := row.Result[2].GetStr()
					msgTime := row.Result[3].GetInt64()
					uid := row.Result[4].GetInt32()
					fileHash := row.Result[5].GetStr()
					msgType := row.Result[6].GetUint32()
					msgSender := row.Result[7].GetStr()

					// 处理 recall 消息
					if msgType >= 16 { // 根据用户定义，type >= 16 表示 recall 消息
						messagesToSend = append(messagesToSend, &pb.ReciveMessageListen{
							Msgid:    int64(msgID),
							Groupid:  int64(groupID),
							Uid:      uid,
							Type:     pb.MessageType(msgType),
							Time:     msgTime,
							Username: msgSender,
							// Msg, Hash, Username 字段在此处留空，因为它们应该为 null
						})
					} else {
						messagesToSend = append(messagesToSend, &pb.ReciveMessageListen{
							Msgid:    int64(msgID),
							Groupid:  int64(groupID),
							Msg:      msgContent,
							Uid:      uid,
							Hash:     fileHash,
							Type:     pb.MessageType(msgType),
							Time:     msgTime,
							Username: msgSender,
						})
					}
				}

				// 将查询结果写入缓存
				cachedList := &pb.CachedMessageList{
					Messages: messagesToSend,
				}
				cacheBytes, err := proto.Marshal(cachedList)
				if err == nil {
					_, err = gateway.ExecRedisBSet(&pbgtw.RedisSetBytesRequest{
						DBID:  0, // 假设 DBID 为 0
						Key:   cacheKey,
						Value: cacheBytes,
						Ttl:   CacheTTL,
					})
				}
			}

			// 如果当前批次没有有效消息，则跳过发送并继续下一轮查询
			if len(messagesToSend) == 0 {
				break // 避免无限循环，如果SQL返回数据但解析失败导致messagesToSend为空
			}

			lastMessageID = messagesToSend[len(messagesToSend)-1].Msgid
			if lastMessageID <= 1 {
				break
			}

			receivedMsgNum += len(messagesToSend)
			if receivedMsgNum > int(req.Limit) {
				messagesToSend = messagesToSend[:req.Limit]
			}

			syncMsg := &pb.SyncMessageResponse{
				Result: &pb.Result{
					Code: errorcode.Success,
					Msg:  "",
				},
				Time: time.Now().Unix(),
				Msg:  messagesToSend,
			}

			// 尝试发送消息，如果失败则重试
			sendSuccess := false
			for range 2 { // 尝试发送2次
				if err := stream.Send(syncMsg); err == nil {
					sendSuccess = true
					break
				}
				time.Sleep(time.Millisecond * 100) // 短暂等待后重试
			}

			if !sendSuccess {
				return fmt.Errorf("failed to send sync message to client")
			}

			if receivedMsgNum >= int(req.Limit) {
				break
			}

		}
	}

	if req.Sync {
		// 等待客户端断开连接
		<-stream.Context().Done()
	}

	return nil
}

// BroadcastMessage 向特定组中的所有客户端发送消息
func BroadcastMessage(groupid int64, msg *pb.SyncMessageResponse) {
	if groupValue, ok := OnlineUsers.Load(groupid); ok {
		groupConns := groupValue.(*GroupConnections)
		groupConns.mu.Lock()
		defer groupConns.mu.Unlock()

		var activeConnections []*ClientConnection
		for _, conn := range groupConns.connections {
			select {
			case conn.sendChan <- msg:
				// 消息已发送到客户端通道
				activeConnections = append(activeConnections, conn)
			default:
				// 客户端通道已满或已断开连接
				close(conn.done)     // 通知发送 goroutine 停止
				close(conn.sendChan) // 关闭发送通道
			}
		}
		groupConns.connections = activeConnections // 更新连接列表
	}
}

// Init 初始化 NATS 订阅
func Init(_ config.Config) error {
	if config.LatestConfig.SyncGRPCProxy.Log {
		log.Printf("[SYNCER]Call Init")
	}
	err := natsmg.Subscribe(natsmg.SubjectMessagePush, func(msg *pb.SyncMessageResponse, natsMsg *nats.Msg) error {
		subjectParts := strings.Split(natsMsg.Subject, ".")
		if len(subjectParts) != 2 || subjectParts[0] != "push" {
			// 无效的 NATS 主题格式
			return fmt.Errorf("invalid NATS subject format")
		}
		groupidStr := subjectParts[1]
		groupid, err := strconv.ParseInt(groupidStr, 10, 64)
		if err != nil {
			// 解析 NATS 主题中的 groupid 错误
			return fmt.Errorf("error parsing groupid from NATS subject")
		}

		// 广播消息给对应组的在线用户
		BroadcastMessage(groupid, msg)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to NATS topic: %w", err)
	}

	return nil
}
