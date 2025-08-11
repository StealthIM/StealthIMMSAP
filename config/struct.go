package config

// Config 主配置
type Config struct {
	DBGateway     DBGatewayConfig     `toml:"dbgateway"`
	SendGRPCProxy SyncGRPCProxyConfig `toml:"send"`
	SyncGRPCProxy SyncGRPCProxyConfig `toml:"sync"`
	Nats          NATSConfig          `toml:"nats"`
	Database      DatabaseConfig      `toml:"database"`
	User          UserConfig          `toml:"user"`
}

// SendGRPCProxyConfig grpc Server配置
type SendGRPCProxyConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
	Log  bool   `toml:"log"`
}

// SyncGRPCProxyConfig grpc Server配置
type SyncGRPCProxyConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
	Log  bool   `toml:"log"`
}

// DBGatewayConfig grpc DBGateway 配置
type DBGatewayConfig struct {
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
	ConnNum int    `toml:"conn_num"`
	Timeout int    `toml:"sql_timeout"`
}

// UserConfig grpc User 配置
type UserConfig struct {
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
	ConnNum int    `toml:"conn_num"`
}

// NATSConfig NATS 消息队列配置
type NATSConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	ConnNum  int    `toml:"conn_num"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	CheckSplitTime int `toml:"check_split_interval"`
	MaxMsgSize     int `toml:"max_msg_size"`
	PartitionSize  int `toml:"partition_size"`
}
