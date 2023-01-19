package etcd

import "context"

type IClient interface {
	// Close 关闭客户端
	Close()
	// Put 保存KV
	Put(key, value string) (*Header, error)
	// PutJson 保存KV（data转成json）
	PutJson(key string, data any) (*Header, error)
	// Get 获取Value值
	Get(key string) (*KeyValue, error)
	// GetPrefixKey 根据KEY前缀获取Value值
	GetPrefixKey(prefixKey string) (map[string]*KeyValue, error)
	// Delete 删除KEY
	Delete(key string) (*Header, error)
	// DeletePrefixKey 根据KEY前缀来删除
	DeletePrefixKey(prefixKey string) (*Header, error)
	// Watch 监听KEY
	Watch(ctx context.Context, key string, watchFunc func(event WatchEvent))
	// WatchPrefixKey 根据KEY前缀来监听
	WatchPrefixKey(ctx context.Context, prefixKey string, watchFunc func(event WatchEvent))
	// Original 原客户端对象
	Original() *etcdClient
}
