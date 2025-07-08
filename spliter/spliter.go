package spliter

import (
	pb "StealthIMMSAP/StealthIM.DBGateway"
	"StealthIMMSAP/config"
	"StealthIMMSAP/errorcode"
	"StealthIMMSAP/gateway"
	"fmt"
	"log"
	"time"
)

// Init 启动分表服务
func Init(cfg config.Config) {
	time.Sleep(10 * time.Second) // 等待数据库初始化完成
	go func() {
		// 首次启动时立即检查一次
		checkAndAddPartitions(cfg)

		ticker := time.NewTicker(time.Duration(cfg.Database.CheckSplitTime) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			checkAndAddPartitions(cfg)
		}
	}()
}

// PartitionInfo 存储分区信息
type PartitionInfo struct {
	Name        string
	Description string // LESS THAN (value) or MAXVALUE
	Value       int64  // 解析后的值，MAXVALUE 为 -1
}

// checkAndAddPartitions 检查并添加或重组新的分区
func checkAndAddPartitions(_ config.Config) {
	log.Println("[SPLITER] Checking partitions")

	// 1. 获取当前 msg 表的最大 msg_id
	sqlReq := &pb.SqlRequest{
		Sql: "SELECT MAX(msg_id) AS `msgid` FROM msg",
		Db:  pb.SqlDatabases_Msg,
	}
	res, err := gateway.ExecSQL(sqlReq)
	if err != nil {
		log.Printf("[SPLITER] Error getting max msg_id: %v", err)
		return
	}

	var maxMsgID int64
	if res.Result.Code == errorcode.Success && len(res.Data) > 0 && len(res.Data[0].Result) > 0 && !res.Data[0].Result[0].GetNull() {
		maxMsgID = int64(res.Data[0].Result[0].GetUint64())
		if maxMsgID == 0 {
			maxMsgID = int64(res.Data[0].Result[0].GetInt64())
		}
	} else {
		// 如果表为空或查询失败，从 0 开始
		maxMsgID = 0
	}

	if maxMsgID == 0 {
		maxMsgID = 1
	}

	targetPartitionsCount := (maxMsgID+int64(config.LatestConfig.Database.PartitionSize)-1)/int64(config.LatestConfig.Database.PartitionSize) + 1

	// 2. 获取当前已存在的分区
	showPartitionsSQL := "SELECT COUNT(*) AS `parts` FROM information_schema.partitions WHERE table_schema = DATABASE() AND table_name = 'msg';"
	partitionsRes, err := gateway.ExecSQL(&pb.SqlRequest{
		Sql: showPartitionsSQL,
		Db:  pb.SqlDatabases_Msg,
	})
	if err != nil {
		log.Printf("[SPLITER] Error getting existing partitions: %v", err)
		return
	}

	if !(partitionsRes.Result.Code == errorcode.Success && len(partitionsRes.Data) > 0) {
		log.Println("[SPLITER] No partitions found in the database")
	}

	partNum := partitionsRes.Data[0].Result[0].GetInt64() - 1

	for i := partNum; i < int64(targetPartitionsCount); i++ {
		// 3. 创建新的分区
		var flag bool = true
		for z := range 3 {
			createPartitionSQL := fmt.Sprintf("ALTER TABLE msg REORGANIZE PARTITION pextra INTO (PARTITION p%d VALUES LESS THAN (%d),PARTITION pextra VALUES LESS THAN MAXVALUE);", i, int64(i+1)*int64(config.LatestConfig.Database.PartitionSize))
			createPartitionRes, err := gateway.ExecSQL(&pb.SqlRequest{
				Sql:    createPartitionSQL,
				Db:     pb.SqlDatabases_Msg,
				Params: []*pb.InterFaceType{},
				Commit: true,
			})
			if err != nil {
				log.Printf("[SPLITER] Error creating partition (%d/3): %v", z, err)
				continue
			}
			if createPartitionRes.Result.Code != errorcode.Success {
				log.Printf("[SPLITER] Error creating partition (%d/3): %v", z, createPartitionRes.Result.Msg)
				continue
			}
			if true {
				flag = false
				break
			}
		}
		if flag {
			return
		}
	}
	log.Println("[SPLITER] Check partitions success")
}
