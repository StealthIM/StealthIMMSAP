package nats

import "github.com/nats-io/nats.go"

// Subject 主题
type Subject int32

const (
	// SubjectMessage 消息
	SubjectMessage Subject = iota
	// SubjectNewTableBroadcast 新建表广播
	SubjectNewTableBroadcast
	// SubjectMessagePush 消息推送
	SubjectMessagePush
	// SubjectCacheInvalidate 缓存失效通知
	SubjectCacheInvalidate
)

var subjectToString = map[Subject]string{
	SubjectMessage:           "message",
	SubjectNewTableBroadcast: "new_table",
	SubjectMessagePush:       "push.*",
	SubjectCacheInvalidate:   "cache.invalidate", // 定义缓存失效主题字符串
}

var subjectQueue = map[Subject]bool{
	SubjectMessage:           true,
	SubjectNewTableBroadcast: false,
	SubjectMessagePush:       true,
	SubjectCacheInvalidate:   false, // 缓存失效通知通常不需要队列
}

var subjectDurable = map[Subject]bool{
	SubjectMessage:           false,
	SubjectNewTableBroadcast: false,
	SubjectMessagePush:       true,
	SubjectCacheInvalidate:   false, // 缓存失效通知通常不需要持久化
}

var streamTable = []nats.StreamConfig{
	{
		Name: "StealthIMMSAP_Messages",
		Subjects: []string{
			subjectToString[SubjectMessage],
		},
	},
	{
		Name: "StealthIMMSAP_NewTableBroadcast",
		Subjects: []string{
			subjectToString[SubjectNewTableBroadcast],
		},
	},
	{
		Name: "StealthIMMSAP_MessagePush",
		Subjects: []string{
			subjectToString[SubjectMessagePush],
		},
	},
	{
		Name: "StealthIMMSAP_CacheInvalidate", // 新增缓存失效流
		Subjects: []string{
			subjectToString[SubjectCacheInvalidate],
		},
	},
}
