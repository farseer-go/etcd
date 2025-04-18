package etcd

import (
	"context"
	"strings"
	"time"

	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/fs/trace"
	etcdV3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var todo = context.TODO()

// etcd客户端别名
type etcdClient = etcdV3.Client

// LeaseID 租约ID
type LeaseID int64

type client struct {
	etcdCli      *etcdClient
	traceManager trace.IManager
}

// 创建客户端
func newClient(config etcdConfig) IClient {
	cli, err := etcdV3.New(etcdV3.Config{
		Endpoints:            strings.Split(config.Server, "|"),
		DialTimeout:          time.Duration(config.DialTimeout) * time.Millisecond,
		DialKeepAliveTime:    time.Duration(config.DialKeepAliveTime) * time.Millisecond,
		DialKeepAliveTimeout: time.Duration(config.DialKeepAliveTimeout) * time.Millisecond,
		MaxCallSendMsgSize:   config.MaxCallSendMsgSize,
		MaxCallRecvMsgSize:   config.MaxCallRecvMsgSize,
		Username:             config.Username,
		Password:             config.Password,
		RejectOldCluster:     config.RejectOldCluster,
		PermitWithoutStream:  config.PermitWithoutStream,
	})
	if err != nil {
		flog.Panic(err)
	}
	return &client{
		etcdCli:      cli,
		traceManager: container.Resolve[trace.IManager](),
	}
}

func (receiver *client) Put(key, value string) (*Header, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("Put", key, 0)
	rsp, err := receiver.etcdCli.Put(todo, key, value)
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return nil, err
	}
	return newResponse(rsp.Header), err
}

func (receiver *client) PutLease(key, value string, leaseId LeaseID) (*Header, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("PutLease", key, int64(leaseId))
	rsp, err := receiver.etcdCli.Put(todo, key, value, etcdV3.WithLease(etcdV3.LeaseID(leaseId)))
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return nil, err
	}
	return newResponse(rsp.Header), err
}

func (receiver *client) PutJson(key string, data any) (*Header, error) {
	jsonValue, _ := snc.Marshal(data)
	return receiver.Put(key, string(jsonValue))
}

func (receiver *client) PutJsonLease(key string, data any, leaseId LeaseID) (*Header, error) {
	jsonValue, _ := snc.Marshal(data)
	return receiver.PutLease(key, string(jsonValue), leaseId)
}

func (receiver *client) Get(key string) (*KeyValue, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("Get", key, 0)

	var result *KeyValue
	rsp, err := receiver.etcdCli.Get(todo, key)
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return nil, err
	}
	if len(rsp.Kvs) > 0 {
		result = newValue(rsp.Kvs[0], rsp.Header)
	} else {
		result = &KeyValue{Header: newResponse(rsp.Header), Key: key}
	}
	return result, err
}

func (receiver *client) GetPrefixKey(prefixKey string) (map[string]*KeyValue, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("GetPrefixKey", prefixKey, 0)

	result := make(map[string]*KeyValue)
	rsp, err := receiver.etcdCli.Get(todo, prefixKey, etcdV3.WithPrefix())
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return result, err
	}
	for _, kv := range rsp.Kvs {
		result[string(kv.Key)] = newValue(kv, rsp.Header)
	}
	return result, err
}

func (receiver *client) Exists(key string) bool {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("Exists", key, 0)

	rsp, err := receiver.etcdCli.Get(todo, key)
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return false
	}
	exists := len(rsp.Kvs) > 0 && rsp.Kvs[0].Version > 0
	return exists
}

func (receiver *client) Delete(key string) (*Header, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("Delete", key, 0)

	rsp, err := receiver.etcdCli.Delete(todo, key)
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return nil, err
	}
	return newResponse(rsp.Header), err
}

func (receiver *client) DeletePrefixKey(prefixKey string) (*Header, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("DeletePrefixKey", prefixKey, 0)

	rsp, err := receiver.etcdCli.Delete(todo, prefixKey, etcdV3.WithPrefix())
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return nil, err
	}
	return newResponse(rsp.Header), err
}

func (receiver *client) Watch(ctx context.Context, key string, watchFunc func(event WatchEvent)) {
	watch := receiver.etcdCli.Watch(ctx, key)
	// 异步处理
	go func() {
		for response := range watch {
			for _, event := range response.Events {
				watchEvent := WatchEvent{
					Type: event.Type.String(),
					Kv:   newValue(event.Kv, &response.Header),
				}
				entryWatchKey := receiver.traceManager.EntryWatchKey(key)
				watchFunc(watchEvent)
				container.Resolve[trace.IManager]().Push(entryWatchKey, nil)
			}
		}
		flog.Info("退出了")
	}()
}

func (receiver *client) WatchPrefixKey(ctx context.Context, prefixKey string, watchFunc func(event WatchEvent)) {
	watch := receiver.etcdCli.Watch(ctx, prefixKey, etcdV3.WithPrefix())
	// 异步处理
	go func() {
		for response := range watch {
			for _, event := range response.Events {
				watchEvent := WatchEvent{
					Type: event.Type.String(),
					Kv:   newValue(event.Kv, &response.Header),
				}
				entryWatchKey := receiver.traceManager.EntryWatchKey(prefixKey)
				watchFunc(watchEvent)
				container.Resolve[trace.IManager]().Push(entryWatchKey, nil)
			}
		}
		flog.Info("退出了")
	}()
}

func (receiver *client) Original() *etcdClient {
	return receiver.etcdCli
}

func (receiver *client) Close() {
	_ = receiver.etcdCli.Close()
}

func (receiver *client) LeaseGrant(ttl int64, keys ...string) (LeaseID, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("LeaseGrant", strings.Join(keys, ","), 0)
	// 生成租约
	leaseGrantResponse, err := receiver.etcdCli.Grant(todo, ttl)
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return 0, err
	}

	// 如果有key，则同时将租约ID赋加到KEY
	if len(keys) > 0 {
		for _, key := range keys {
			// 是否有更简单的方案，不需要使用PUT，而直接对KEY赋加
			rsp, err := receiver.etcdCli.Get(todo, key)
			if err != nil {
				continue
			}
			if len(rsp.Kvs) > 0 {
				_, err = receiver.etcdCli.Put(todo, key, string(rsp.Kvs[0].Value), etcdV3.WithLease(leaseGrantResponse.ID))
				flog.ErrorIfExists(err)
			}
		}
	}
	return LeaseID(leaseGrantResponse.ID), nil
}

func (receiver *client) LeaseKeepAlive(ctx context.Context, leaseId LeaseID) error {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("LeaseKeepAlive", "", int64(leaseId))
	keepRespChan, err := receiver.etcdCli.KeepAlive(ctx, etcdV3.LeaseID(leaseId))
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return err
	}

	// 自动续租
	go func() {
		for {
			select {
			case keepResp := <-keepRespChan:
				if keepRespChan == nil {
					flog.Debugf("续租解除：%d", keepResp.ID)
					return
				} else { // 每秒会续租一次, 所以就会受到一次应答
					flog.Debugf("续租成功：%d", keepResp.ID)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return err
}

func (receiver *client) LeaseKeepAliveOnce(leaseId LeaseID) error {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("LeaseKeepAliveOnce", "", int64(leaseId))
	_, err := receiver.etcdCli.KeepAliveOnce(todo, etcdV3.LeaseID(leaseId))
	defer func() { traceDetailEtcd.End(err) }()

	return err
}

func (receiver *client) LeaseRevoke(leaseId LeaseID) (*Header, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("LeaseRevoke", "", int64(leaseId))
	revoke, err := receiver.etcdCli.Revoke(todo, etcdV3.LeaseID(leaseId))
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return nil, err
	}
	return newResponse(revoke.Header), nil
}

func (receiver *client) LeaseInfo(leaseId LeaseID) (*LeaseInfo, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("LeaseInfo", "", int64(leaseId))

	leaseTimeToLiveResponse, err := receiver.etcdCli.TimeToLive(todo, etcdV3.LeaseID(leaseId))
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return nil, err
	}
	return newLeaseInfo(leaseTimeToLiveResponse), nil
}

// UnLock 解锁
type UnLock func()

func (receiver *client) Lock(lockKey string, lockTTL int) (UnLock, error) {
	traceDetailEtcd := receiver.traceManager.TraceEtcd("Lock", lockKey, 0)

	ctx, cancelFunc := context.WithCancel(context.Background())
	session, err := concurrency.NewSession(receiver.etcdCli, concurrency.WithTTL(lockTTL))
	defer func() { traceDetailEtcd.End(err) }()

	if err != nil {
		return func() { cancelFunc() }, err
	}
	mutex := concurrency.NewMutex(session, lockKey)
	err = mutex.Lock(ctx)
	if err != nil {

		return func() { cancelFunc() }, err
	}

	return func() {
		_ = mutex.Unlock(ctx)
		cancelFunc()
	}, nil
}
