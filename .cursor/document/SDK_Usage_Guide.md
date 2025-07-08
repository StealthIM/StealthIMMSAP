# API 使用指南 (AI 视角)

本文档总结了 `nats/` 和 `gateway/` 目录中提供的核心 API 函数的使用方法，旨在帮助 AI 理解如何调用这些接口以完成特定任务。

**注意**: 在本指南中，`pb` 是一个常用的别名，用于导入 Protobuf 定义包。例如，`*pb.RedisGetStringRequest` 实际上指的是 `*StealthIM.DBGateway.RedisGetStringRequest`，而 `*pb.SendMessageRequest` 则指的是 `*StealthIM.MSAP.SendMessageRequest`。具体 `pb` 别名代表哪个包取决于其在代码中的导入声明。

**注意**: 如果在一个 go 中文件需要使用多个不同的包，`pb` 应 改为如 `pbgtw`、`pbmsap` 这类全小写的使用缩写的别名。当前模块所对应的 proto 必须使用 `pb`。

## `nats/` 目录 API

`nats/` 目录提供了与 NATS 消息队列系统交互的功能，用于发布和订阅消息。

### `nats.Publish(subject Subject, data proto.Message) error`

* **功能**: 使用 NATS JetStream 发布 Protobuf 消息到指定主题。
* **参数**:
  * `subject`: `nats.Subject` 类型，表示预定义的主题（例如 `nats.SubjectMessage`）。
  * `data`: `proto.Message` 接口类型，表示要发布的 Protobuf 消息数据。
* **返回**: `error`，如果发布成功则为 `nil`，否则为错误信息。
* **AI 使用场景**: 当需要向 NATS JetStream 发送结构化数据（Protobuf 消息）时使用，例如发送聊天消息、系统通知等。

**相关 Protobuf 定义**:

```protobuf
// proto/msap.proto
message SendMessageRequest {
  int32 uid = 1;        // 发送者 uid
  int64 groupid = 2;    // 群 id
  string msg = 3;       // 消息内容
  MessageType type = 4; // 消息类型
}

enum MessageType { // 消息类型
  Text = 0;        // 文字
  Image = 1;       // 纯图片
  LargeEmoji = 2;  // 大表情
  Emoji = 3;       // 表情
  File = 4;        // 文件
  Card = 5;        // 卡片
  InnerLink = 6;   // 内部链接
  // 以下为撤回消息类型
  Recall_Text = 16;
  Recall_Image = 17;
  Recall_LargeEmoji = 18;
  Recall_Emoji = 19;
  Recall_File = 20;
  Recall_Card = 21;
  Recall_InnerLink = 22;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/nats"
 pb "StealthIMMSAP/StealthIM.MSAP" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 NATS 连接已初始化
 // nats.InitConns() // 在实际应用中，这会在后台运行

 req := &pb.SendMessageRequest{
  Uid:     123,
  Groupid: 456,
  Msg:     "Hello, NATS JetStream!",
  Type:    pb.MessageType_Text,
 }

 err := nats.Publish(nats.SubjectMessage, req)
 if err != nil {
  log.Printf("Failed to publish message: %v", err)
 } else {
  log.Println("Message published successfully to JetStream.")
 }
}
```

### `nats.PublishWithSubject(subject string, data proto.Message) error`

* **功能**: 使用 NATS JetStream 发布 Protobuf 消息到指定字符串主题。
* **参数**:
  * `subject`: `string` 类型，表示自定义的主题字符串。
  * `data`: `proto.Message` 接口类型，表示要发布的 Protobuf 消息数据。
* **返回**: `error`，如果发布成功则为 `nil`，否则为错误信息。
* **AI 使用场景**: 当需要向 NATS JetStream 发送结构化数据（Protobuf 消息）到动态或非预定义主题时使用。

**相关 Protobuf 定义**:

```protobuf
// proto/msap.proto
message SendMessageRequest {
  int32 uid = 1;        // 发送者 uid
  int64 groupid = 2;    // 群 id
  string msg = 3;       // 消息内容
  MessageType type = 4; // 消息类型
}

enum MessageType { // 消息类型
  Text = 0;        // 文字
  Image = 1;       // 纯图片
  LargeEmoji = 2;  // 大表情
  Emoji = 3;       // 表情
  File = 4;        // 文件
  Card = 5;        // 卡片
  InnerLink = 6;   // 内部链接
  // 以下为撤回消息类型
  Recall_Text = 16;
  Recall_Image = 17;
  Recall_LargeEmoji = 18;
  Recall_Emoji = 19;
  Recall_File = 20;
  Recall_Card = 21;
  Recall_InnerLink = 22;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/nats"
 pb "StealthIMMSAP/StealthIM.MSAP" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 NATS 连接已初始化
 // nats.InitConns() // 在实际应用中，这会在后台运行

 req := &pb.SendMessageRequest{
  Uid:     123,
  Groupid: 456,
  Msg:     "Hello, NATS JetStream with custom subject!",
  Type:    pb.MessageType_Text,
 }

 err := nats.PublishWithSubject("my.custom.subject", req)
 if err != nil {
  log.Printf("Failed to publish message with custom subject: %v", err)
 } else {
  log.Println("Message published successfully to JetStream with custom subject.")
 }
}
```

### `nats.PublishCore(subject Subject, data proto.Message) error`

* **功能**: 使用 NATS Core 发布 Protobuf 消息到指定主题。
* **参数**:
  * `subject`: `nats.Subject` 类型，表示预定义的主题。
  * `data`: `proto.Message` 接口类型，表示要发布的 Protobuf 消息数据。
* **返回**: `error`，如果发布成功则为 `nil`，否则为错误信息。
* **AI 使用场景**: 当需要向 NATS Core 发送结构化数据（Protobuf 消息）且不需要 JetStream 的持久化和流特性时使用。

**相关 Protobuf 定义**:

```protobuf
// proto/msap.proto
message SendMessageRequest {
  int32 uid = 1;        // 发送者 uid
  int64 groupid = 2;    // 群 id
  string msg = 3;       // 消息内容
  MessageType type = 4; // 消息类型
}

enum MessageType { // 消息类型
  Text = 0;        // 文字
  Image = 1;       // 纯图片
  LargeEmoji = 2;  // 大表情
  Emoji = 3;       // 表情
  File = 4;        // 文件
  Card = 5;        // 卡片
  InnerLink = 6;   // 内部链接
  // 以下为撤回消息类型
  Recall_Text = 16;
  Recall_Image = 17;
  Recall_LargeEmoji = 18;
  Recall_Emoji = 19;
  Recall_File = 20;
  Recall_Card = 21;
  Recall_InnerLink = 22;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/nats"
 pb "StealthIMMSAP/StealthIM.MSAP" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 NATS 连接已初始化
 // nats.InitConns() // 在实际应用中，这会在后台运行

 req := &pb.SendMessageRequest{
  Uid:     123,
  Groupid: 456,
  Msg:     "Hello, NATS Core!",
  Type:    pb.MessageType_Text,
 }

 err := nats.PublishCore(nats.SubjectMessage, req)
 if err != nil {
  log.Printf("Failed to publish message to NATS Core: %v", err)
 } else {
  log.Println("Message published successfully to NATS Core.")
 }
}
```

### `nats.Subscribe[msgType proto.Message](subject Subject, handler func(msgType, *nats.Msg) error) error`

* **功能**: 订阅 NATS JetStream 消息。支持队列订阅和非队列订阅。
* **参数**:
  * `subject`: `nats.Subject` 类型，表示要订阅的主题。
  * `handler`: 一个回调函数，当收到消息时被调用。
    * `msgType`: 泛型类型，表示 Protobuf 消息的具体类型，它必须实现 `proto.Message` 接口（例如 `*StealthIM.MSAP.ReciveMessageListen`）。
    * `*nats.Msg`: 原始的 NATS 消息对象。
* **返回**: `error`，如果订阅成功则为 `nil`，否则为错误信息。
* **AI 使用场景**: 当需要从 NATS JetStream 接收特定类型的 Protobuf 消息并进行处理时使用，例如处理传入的聊天消息、广播通知等。

**相关 Protobuf 定义**:

```protobuf
// proto/msap.proto
message ReciveMessageListen {
  int64 msgid = 1;      // 消息id
  int32 uid = 2;        // 发送者 uid
  int64 groupid = 3;    // 群 id
  string msg = 4;       // 消息内容（文件名）
  MessageType type = 5; // 消息类型
  string hash = 6;      // 文件hash
  string username = 7;  // 发送者用户名（减少二次查表）
}

enum MessageType { // 消息类型
  Text = 0;        // 文字
  Image = 1;       // 纯图片
  LargeEmoji = 2;  // 大表情
  Emoji = 3;       // 表情
  File = 4;        // 文件
  Card = 5;        // 卡片
  InnerLink = 6;   // 内部链接
  // 以下为撤回消息类型
  Recall_Text = 16;
  Recall_Image = 17;
  Recall_LargeEmoji = 18;
  Recall_Emoji = 19;
  Recall_File = 20;
  Recall_Card = 21;
  Recall_InnerLink = 22;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/nats"
 pb "StealthIMMSAP/StealthIM.MSAP" // 导入对应的 Protobuf 包
 "log"

 "github.com/nats-io/nats.go"
)

func main() {
 // 假设 NATS 连接已初始化
 // nats.InitConns() // 在实际应用中，这会在后台运行

 err := nats.Subscribe(nats.SubjectMessage, func(msg *pb.ReciveMessageListen, natsMsg *nats.Msg) error {
  log.Printf("Received JetStream message: MsgID=%d, UID=%d, GroupID=%d, Msg='%s', Type=%s, Hash='%s', Username='%s'",
   msg.Msgid, msg.Uid, msg.Groupid, msg.Msg, msg.Type.String(), msg.Hash, msg.Username)
  natsMsg.Ack() // 确认消息
  return nil
 })

 if err != nil {
  log.Printf("Failed to subscribe to JetStream messages: %v", err)
 } else {
  log.Println("Successfully subscribed to JetStream messages. Waiting for messages...")
  // 在实际应用中，这里可能需要一个阻塞操作来保持主goroutine运行，例如 select {}
 }
}
```

### `nats.SubscribeCore[msgType proto.Message](subject Subject, handler func(msgType)) error`

* **功能**: 订阅 NATS Core 消息。支持队列订阅和非队列订阅。
* **参数**:
  * `subject`: `nats.Subject` 类型，表示要订阅的主题。
  * `handler`: 一个回调函数，当收到消息时被调用。
    * `msgType`: 泛型类型，表示 Protobuf 消息的具体类型，它必须实现 `proto.Message` 接口（例如 `*StealthIM.MSAP.ReciveMessageListen`）。
* **返回**: `error`，如果订阅成功则为 `nil`，否则为错误信息。
* **AI 使用场景**: 当需要从 NATS Core 接收特定类型的 Protobuf 消息且不需要 JetStream 特性时使用。

**相关 Protobuf 定义**:

```protobuf
// proto/msap.proto
message ReciveMessageListen {
  int64 msgid = 1;      // 消息id
  int32 uid = 2;        // 发送者 uid
  int64 groupid = 3;    // 群 id
  string msg = 4;       // 消息内容（文件名）
  MessageType type = 5; // 消息类型
  string hash = 6;      // 文件hash
  string username = 7;  // 发送者用户名（减少二次查表）
}

enum MessageType { // 消息类型
  Text = 0;        // 文字
  Image = 1;       // 纯图片
  LargeEmoji = 2;  // 大表情
  Emoji = 3;       // 表情
  File = 4;        // 文件
  Card = 5;        // 卡片
  InnerLink = 6;   // 内部链接
  // 以下为撤回消息类型
  Recall_Text = 16;
  Recall_Image = 17;
  Recall_LargeEmoji = 18;
  Recall_Emoji = 19;
  Recall_File = 20;
  Recall_Card = 21;
  Recall_InnerLink = 22;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/nats"
 pb "StealthIMMSAP/StealthIM.MSAP" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 NATS 连接已初始化
 // nats.InitConns() // 在实际应用中，这会在后台运行

 err := nats.SubscribeCore(nats.SubjectMessage, func(msg *pb.ReciveMessageListen) {
  log.Printf("Received NATS Core message: MsgID=%d, UID=%d, GroupID=%d, Msg='%s', Type=%s, Hash='%s', Username='%s'",
   msg.Msgid, msg.Uid, msg.Groupid, msg.Msg, msg.Type.String(), msg.Hash, msg.Username)
 })

 if err != nil {
  log.Printf("Failed to subscribe to NATS Core messages: %v", err)
 } else {
  log.Println("Successfully subscribed to NATS Core messages. Waiting for messages...")
  // 在实际应用中，这里可能需要一个阻塞操作来保持主goroutine运行，例如 select {}
 }
}
```

## `gateway/` 目录 API

`gateway/` 目录提供了与 DBGateway gRPC 服务交互的功能，用于执行数据库操作。

### `gateway.ExecRedisGet(req *pb.RedisGetStringRequest) (*pb.RedisGetStringResponse, error)`

* **功能**: 执行 Redis GET 命令，获取字符串类型的值。
* **参数**:
  * `req`: `*pb.RedisGetStringRequest` 类型，包含要查询的键。
* **返回**: `*pb.RedisGetStringResponse` (包含查询结果) 和 `error`。
* **AI 使用场景**: 从 Redis 缓存中读取字符串数据，例如用户会话信息、配置参数等。

**相关 Protobuf 定义**:

```protobuf
// proto/db_gateway.proto
message RedisGetStringRequest {
  int32 DBID = 1;
  string key = 2;
}
message RedisGetStringResponse {
  int32 DBID = 1;
  Result result = 2;
  string value = 3;
}
message Result {
  int32 code = 1;
  string msg = 2;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/gateway"
 pb "StealthIMMSAP/StealthIM.DBGateway" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 DBGateway 连接已初始化
 // gateway.InitConns() // 在实际应用中，这会在后台运行

 req := &pb.RedisGetStringRequest{
  DBID: 0, // 示例数据库ID
  Key:  "my_string_key",
 }

 res, err := gateway.ExecRedisGet(req)
 if err != nil {
  log.Printf("Failed to get Redis string: %v", err)
 } else {
  if res.Result.Code == 800 { // 假设 800 是成功代码
   log.Printf("Successfully got Redis string: Key='%s', Value='%s'", req.Key, res.Value)
  } else {
   log.Printf("Failed to get Redis string: Code=%d, Msg='%s'", res.Result.Code, res.Result.Msg)
  }
 }
}
```

### `gateway.ExecRedisSet(req *pb.RedisSetStringRequest) (*pb.RedisSetResponse, error)`

* **功能**: 执行 Redis SET 命令，设置字符串类型的值。
* **参数**:
  * `req`: `*pb.RedisSetStringRequest` 类型，包含要设置的键、值和过期时间。
* **返回**: `*pb.RedisSetResponse` (包含操作结果) 和 `error`。
* **AI 使用场景**: 向 Redis 缓存写入字符串数据，例如存储用户会话、临时数据等。

**相关 Protobuf 定义**:

```protobuf
// proto/db_gateway.proto
message RedisSetStringRequest {
  int32 DBID = 1;
  string key = 2;
  string value = 3;
  int32 ttl = 4;
}
message RedisSetResponse { Result result = 1; }
message Result {
  int32 code = 1;
  string msg = 2;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/gateway"
 pb "StealthIMMSAP/StealthIM.DBGateway" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 DBGateway 连接已初始化
 // gateway.InitConns() // 在实际应用中，这会在后台运行

 req := &pb.RedisSetStringRequest{
  DBID:  0, // 示例数据库ID
  Key:   "my_new_string_key",
  Value: "some_value",
  Ttl:   60, // 60秒过期
 }

 res, err := gateway.ExecRedisSet(req)
 if err != nil {
  log.Printf("Failed to set Redis string: %v", err)
 } else {
  if res.Result.Code == 800 { // 假设 800 是成功代码
   log.Printf("Successfully set Redis string: Key='%s', Value='%s'", req.Key, req.Value)
  } else {
   log.Printf("Failed to set Redis string: Code=%d, Msg='%s'", res.Result.Code, res.Result.Msg)
  }
 }
}
```

### `gateway.ExecRedisBGet(req *pb.RedisGetBytesRequest) (*pb.RedisGetBytesResponse, error)`

* **功能**: 执行 Redis GET 命令，获取二进制类型的值。
* **参数**:
  * `req`: `*pb.RedisGetBytesRequest` 类型，包含要查询的键。
* **返回**: `*pb.RedisGetBytesResponse` (包含查询结果) 和 `error`。
* **AI 使用场景**: 从 Redis 缓存中读取二进制数据，例如序列化的对象、图片等。

**相关 Protobuf 定义**:

```protobuf
// proto/db_gateway.proto
message RedisGetBytesRequest {
  int32 DBID = 1;
  string key = 2;
}
message RedisGetBytesResponse {
  int32 DBID = 1;
  Result result = 2;
  bytes value = 3;
}
message Result {
  int32 code = 1;
  string msg = 2;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/gateway"
 pb "StealthIMMSAP/StealthIM.DBGateway" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 DBGateway 连接已初始化
 // gateway.InitConns() // 在实际应用中，这会在后台运行

 req := &pb.RedisGetBytesRequest{
  DBID: 0, // 示例数据库ID
  Key:  "my_binary_key",
 }

 res, err := gateway.ExecRedisBGet(req)
 if err != nil {
  log.Printf("Failed to get Redis binary: %v", err)
 } else {
  if res.Result.Code == 800 { // 假设 800 是成功代码
   log.Printf("Successfully got Redis binary: Key='%s', ValueLength=%d", req.Key, len(res.Value))
  } else {
   log.Printf("Failed to get Redis binary: Code=%d, Msg='%s'", res.Result.Code, res.Result.Msg)
  }
 }
}
```

### `gateway.ExecRedisBSet(req *pb.RedisSetBytesRequest) (*pb.RedisSetResponse, error)`

* **功能**: 执行 Redis SET 命令，设置二进制类型的值。
* **参数**:
  * `req`: `*pb.RedisSetBytesRequest` 类型，包含要设置的键、值和过期时间。
* **返回**: `*pb.RedisSetResponse` (包含操作结果) 和 `error`。
* **AI 使用场景**: 向 Redis 缓存写入二进制数据，例如存储序列化的对象、图片等。

**相关 Protobuf 定义**:

```protobuf
// proto/db_gateway.proto
message RedisSetBytesRequest {
  int32 DBID = 1;
  string key = 2;
  bytes value = 3;
  int32 ttl = 4;
}
message RedisSetResponse { Result result = 1; }
message Result {
  int32 code = 1;
  string msg = 2;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/gateway"
 pb "StealthIMMSAP/StealthIM.DBGateway" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 DBGateway 连接已初始化
 // gateway.InitConns() // 在实际应用中，这会在后台运行

 req := &pb.RedisSetBytesRequest{
  DBID:  0, // 示例数据库ID
  Key:   "my_new_binary_key",
  Value: []byte("some binary data"),
  Ttl:   60, // 60秒过期
 }

 res, err := gateway.ExecRedisBSet(req)
 if err != nil {
  log.Printf("Failed to set Redis binary: %v", err)
 } else {
  if res.Result.Code == 800 { // 假设 800 是成功代码
   log.Printf("Successfully set Redis binary: Key='%s', ValueLength=%d", req.Key, len(req.Value))
  } else {
   log.Printf("Failed to set Redis binary: Code=%d, Msg='%s'", res.Result.Code, res.Result.Msg)
  }
 }
}
```

### `gateway.ExecRedisDel(req *pb.RedisDelRequest) (*pb.RedisDelResponse, error)`

* **功能**: 执行 Redis DEL 命令，删除一个或多个键。
* **参数**:
  * `req`: `*pb.RedisDelRequest` 类型，包含要删除的键列表。
* **返回**: `*pb.RedisDelResponse` (包含删除的键数量) 和 `error`。
* **AI 使用场景**: 从 Redis 缓存中删除不再需要的数据。

**相关 Protobuf 定义**:

```protobuf
// proto/db_gateway.proto
message RedisDelRequest {
  int32 DBID = 1;
  string key = 2;
}
message RedisDelResponse { Result result = 1; }
message Result {
  int32 code = 1;
  string msg = 2;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/gateway"
 pb "StealthIMMSAP/StealthIM.DBGateway" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 DBGateway 连接已初始化
 // gateway.InitConns() // 在实际应用中，这会在后台运行

 req := &pb.RedisDelRequest{
  DBID: 0, // 示例数据库ID
  Key:  "my_key_to_delete",
 }

 res, err := gateway.ExecRedisDel(req)
 if err != nil {
  log.Printf("Failed to delete Redis key: %v", err)
 } else {
  if res.Result.Code == 800 { // 假设 800 是成功代码
   log.Printf("Successfully deleted Redis key: Key='%s'", req.Key)
  } else {
   log.Printf("Failed to delete Redis key: Code=%d, Msg='%s'", res.Result.Code, res.Result.Msg)
  }
 }
}
```

### `gateway.ExecSQL(sql *pb.SqlRequest) (*pb.SqlResponse, error)`

* **功能**: 执行 SQL 语句。
* **参数**:
  * `sql`: `*pb.SqlRequest` 类型，包含要执行的 SQL 语句和参数。
* **返回**: `*pb.SqlResponse` (包含查询结果或受影响的行数) 和 `error`。
* **AI 使用场景**: 执行数据库查询、插入、更新或删除操作。

**相关 Protobuf 定义**:

```protobuf
// proto/db_gateway.proto
message SqlRequest {
  string sql = 1;
  SqlDatabases db = 2;
  repeated InterFaceType params = 3;
  bool commit = 4;
  bool get_row_count = 5;
  bool get_last_insert_id = 6;
}

enum SqlDatabases {
  Users = 0;
  Msg = 1;
  File = 2;
  Logging = 3;
  Groups = 4;
  Masterdb = 5;
  Session = 6;
};

message InterFaceType {
  oneof response {
    string str = 1;
    int32 int32 = 2;
    int64 int64 = 3;
    bool bool = 4;
    float float = 5;
    double double = 6;
    bytes blob = 7;
  }
  bool null = 8;
}

message SqlLine { repeated InterFaceType result = 1; }

message SqlResponse {
  Result result = 1;
  int64 rows_affected = 2;
  int64 last_insert_id = 3;
  repeated SqlLine data = 4;
}
message Result {
  int32 code = 1;
  string msg = 2;
}
```

**Go 调用示例**:

```go
package main

import (
 "StealthIMMSAP/gateway"
 pb "StealthIMMSAP/StealthIM.DBGateway" // 导入对应的 Protobuf 包
 "log"
)

func main() {
 // 假设 DBGateway 连接已初始化
 // gateway.InitConns() // 在实际应用中，这会在后台运行

 // 示例：插入数据
 insertReq := &pb.SqlRequest{
  Sql:    "INSERT INTO users (username, nickname) VALUES (?, ?)",
  Db:     pb.SqlDatabases_Users,
  Params: []*pb.InterFaceType{{Str: "testuser"}, {Str: "Test Nickname"}},
  Commit: true,
  GetLastInsertId: true,
 }

 insertRes, err := gateway.ExecSQL(insertReq)
 if err != nil {
  log.Printf("Failed to insert user: %v", err)
 } else {
  if insertRes.Result.Code == 800 {
   log.Printf("Successfully inserted user. Last Insert ID: %d", insertRes.LastInsertId)
  } else {
   log.Printf("Failed to insert user: Code=%d, Msg='%s'", insertRes.Result.Code, insertRes.Result.Msg)
  }
 }

 // 示例：查询数据
 queryReq := &pb.SqlRequest{
  Sql:    "SELECT username, nickname FROM users WHERE username = ?",
  Db:     pb.SqlDatabases_Users,
  Params: []*pb.InterFaceType{{Str: "testuser"}},
 }

 queryRes, err := gateway.ExecSQL(queryReq)
 if err != nil {
  log.Printf("Failed to query user: %v", err)
 } else {
  if queryRes.Result.Code == 800 {
   log.Printf("Successfully queried user. Rows: %d", len(queryRes.Data))
   for _, row := range queryRes.Data {
    log.Printf("  Username: %s, Nickname: %s", row.Result[0].Str, row.Result[1].Str)
   }
  } else {
   log.Printf("Failed to query user: Code=%d, Msg='%s'", queryRes.Result.Code, queryRes.Result.Msg)
  }
 }
}
```

! 无论如何你应该优先使用上述 sdk 而不是直接链接！
! 无论如何你应该优先使用上述 sdk 而不是直接链接！
! 无论如何你应该优先使用上述 sdk 而不是直接链接！
