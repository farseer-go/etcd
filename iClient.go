package etcd

type IClient interface {
	// Close 关闭客户端
	Close()
	// Put 保存KV
	Put(key, value string) (*Response, error)
	// PutJson 保存KV（data转成json）
	PutJson(key string, data any) (*Response, error)
	// Get 获取Value值
	Get(key string) (*KeyValue, error)
	// GetPrefixKey 根据KEY前缀获取Value值
	GetPrefixKey(prefixKey string) (map[string]*KeyValue, error)
	// Delete 删除KEY
	Delete(key string) (*Response, error)
	// DeletePrefixKey 根据KEY前缀来删除
	DeletePrefixKey(prefixKey string) (*Response, error)
	// Original 原客户端对象
	Original() *etcdClient
}
