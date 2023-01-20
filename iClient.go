package etcd

import "context"

type IClient interface {
	// Close 关闭客户端
	Close()
	// Put 保存KV
	Put(key, value string) (*Header, error)
	// PutLease 保存KV，同时赋加租约
	PutLease(key, value string, leaseId LeaseID) (*Header, error)
	// PutJson 保存KV（data转成json）
	PutJson(key string, data any) (*Header, error)
	// PutJsonLease 保存KV（data转成json），同时赋加租约
	PutJsonLease(key string, data any, leaseId LeaseID) (*Header, error)
	// Get 获取Value值
	Get(key string) (*KeyValue, error)
	// GetPrefixKey 根据KEY前缀获取Value值
	GetPrefixKey(prefixKey string) (map[string]*KeyValue, error)
	// Delete 删除KEY
	Delete(key string) (*Header, error)
	// DeletePrefixKey 根据KEY前缀来删除
	DeletePrefixKey(prefixKey string) (*Header, error)
	// Exists 判断是否存在KEY
	Exists(key string) (bool, error)
	// Watch 监听KEY
	Watch(ctx context.Context, key string, watchFunc func(event WatchEvent))
	// WatchPrefixKey 根据KEY前缀来监听
	WatchPrefixKey(ctx context.Context, prefixKey string, watchFunc func(event WatchEvent))
	// LeaseGrant 创建租约，ttl：租约的时间（单位s），keys：要赋加租约的KEY
	LeaseGrant(ttl int64, keys ...string) (LeaseID, error)
	// LeaseKeepAlive 续租（持续）
	LeaseKeepAlive(ctx context.Context, leaseId LeaseID) error
	// LeaseKeepAliveOnce 续租一次
	LeaseKeepAliveOnce(leaseId LeaseID) error
	// LeaseRevoke 删除租约（会使当前租约的所关联的key-value失效）
	LeaseRevoke(leaseId LeaseID) (*Header, error)
	// LeaseInfo 查询租约信息
	LeaseInfo(leaseId LeaseID) (*LeaseInfo, error)
	// Lock 添加锁
	Lock(lockKey string, lockTTL int) (UnLock, error)
	// Original 原客户端对象
	Original() *etcdClient
}
