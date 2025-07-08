package nats

import "github.com/nats-io/nats.go"

// Subject 主题
type Subject int32

const (
	// SubjectMessage 消息
	SubjectMessage Subject = iota
	// SubjectMessagePush 消息推送
	SubjectMessagePush
	// SubjectCacheInvalidate 缓存失效通知
	SubjectCacheInvalidate
)

var subjectToString = map[Subject]string{
	SubjectMessage:         "message",
	SubjectMessagePush:     "push.*",
	SubjectCacheInvalidate: "cache.invalidate", // 定义缓存失效主题字符串
}

var subjectQueue = map[Subject]bool{
	SubjectMessage:         true,
	SubjectMessagePush:     true,
	SubjectCacheInvalidate: true,
}

var subjectDurable = map[Subject]bool{
	SubjectMessage:         false,
	SubjectMessagePush:     true,
	SubjectCacheInvalidate: false,
}

var streamTable = []nats.StreamConfig{
	{
		Name: "StealthIMMSAP_Messages",
		Subjects: []string{
			subjectToString[SubjectMessage],
		},
	},
	{
		Name: "StealthIMMSAP_MessagePush",
		Subjects: []string{
			subjectToString[SubjectMessagePush],
		},
	},
	{
		Name: "StealthIMMSAP_CacheInvalidate", // 缓存失效流
		Subjects: []string{
			subjectToString[SubjectCacheInvalidate],
		},
	},
}
