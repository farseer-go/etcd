package etcd

type etcdConfig struct {
	Server               string // 服务端地址
	DialTimeout          int    // 连接超时时间（ms)
	DialKeepAliveTime    int    // 对服务器进行ping的时间（ms)
	DialKeepAliveTimeout int    // 客户端等待响应的超时时间（ms)
	MaxCallSendMsgSize   int    // 客户端的请求发送限制，单位是字节。0，则默认为2.0 MiB（2 * 1024 * 1024）。
	MaxCallRecvMsgSize   int    // 客户端的响应接收限制，单位是字节。0，则默认不限制
	Username             string // 用户名
	Password             string // 密码
	RejectOldCluster     bool   // 拒绝过时的集群创建客户端。
	PermitWithoutStream  bool   // 允许客户端在没有任何活动流（RPC）的情况下向服务器发送keepalive pings。
}
