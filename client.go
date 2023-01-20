package etcd

import (
	"context"
	"encoding/json"
	"github.com/farseer-go/fs/flog"
	etcdV3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"strings"
	"time"
)

var todo = context.TODO()

// etcd客户端别名
type etcdClient = etcdV3.Client

// LeaseID 租约ID
type LeaseID int64

type client struct {
	etcdCli *etcdClient
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
		etcdCli: cli,
	}
}

func (receiver *client) Put(key, value string) (*Header, error) {
	rsp, err := receiver.etcdCli.Put(todo, key, value)
	if err != nil {
		_ = flog.Error(err)
		return nil, err
	}
	return newResponse(rsp.Header), err
}

func (receiver *client) PutLease(key, value string, leaseId LeaseID) (*Header, error) {
	rsp, err := receiver.etcdCli.Put(todo, key, value, etcdV3.WithLease(etcdV3.LeaseID(leaseId)))
	if err != nil {
		_ = flog.Error(err)
		return nil, err
	}
	return newResponse(rsp.Header), err
}

func (receiver *client) PutJson(key string, data any) (*Header, error) {
	jsonValue, _ := json.Marshal(data)
	return receiver.Put(key, string(jsonValue))
}

func (receiver *client) PutJsonLease(key string, data any, leaseId LeaseID) (*Header, error) {
	jsonValue, _ := json.Marshal(data)
	return receiver.PutLease(key, string(jsonValue), leaseId)
}

func (receiver *client) Get(key string) (*KeyValue, error) {
	var result *KeyValue
	rsp, err := receiver.etcdCli.Get(todo, key)
	if err != nil {
		_ = flog.Error(err)
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
	result := make(map[string]*KeyValue)
	rsp, err := receiver.etcdCli.Get(todo, prefixKey, etcdV3.WithPrefix())
	if err != nil {
		_ = flog.Error(err)
		return result, err
	}
	for _, kv := range rsp.Kvs {
		result[string(kv.Key)] = newValue(kv, rsp.Header)
	}
	return result, err
}

func (receiver *client) Exists(key string) bool {
	rsp, err := receiver.etcdCli.Get(todo, key)
	if err != nil {
		_ = flog.Error(err)
		return false
	}
	exists := len(rsp.Kvs) > 0 && rsp.Kvs[0].Version > 0
	return exists
}

func (receiver *client) Delete(key string) (*Header, error) {
	rsp, err := receiver.etcdCli.Delete(todo, key)
	if err != nil {
		_ = flog.Error(err)
		return nil, err
	}
	return newResponse(rsp.Header), err
}

func (receiver *client) DeletePrefixKey(prefixKey string) (*Header, error) {
	rsp, err := receiver.etcdCli.Delete(todo, prefixKey, etcdV3.WithPrefix())
	if err != nil {
		_ = flog.Error(err)
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
				watchFunc(watchEvent)
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
				watchFunc(watchEvent)
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
	// 生成租约
	leaseGrantResponse, err := receiver.etcdCli.Grant(todo, ttl)
	if err != nil {
		_ = flog.Error(err)
		return 0, err
	}

	// 如果有key，则同时将租约ID赋加到KEY
	if len(keys) > 0 {
		for _, key := range keys {
			// 是否有更简单的方案，不需要使用PUT，而直接对KEY赋加
			rsp, err := receiver.etcdCli.Get(todo, key)
			if err != nil {
				_ = flog.Error(err)
				continue
			}
			if len(rsp.Kvs) > 0 {
				_, err = receiver.etcdCli.Put(todo, key, string(rsp.Kvs[0].Value), etcdV3.WithLease(leaseGrantResponse.ID))
				if err != nil {
					_ = flog.Error(err)
				}
			}
		}
	}
	return LeaseID(leaseGrantResponse.ID), nil
}

func (receiver *client) LeaseKeepAlive(ctx context.Context, leaseId LeaseID) error {
	keepRespChan, err := receiver.etcdCli.KeepAlive(ctx, etcdV3.LeaseID(leaseId))
	if err != nil {
		_ = flog.Error(err)
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
	_, err := receiver.etcdCli.KeepAliveOnce(todo, etcdV3.LeaseID(leaseId))
	if err != nil {
		_ = flog.Error(err)
	}
	return err
}

func (receiver *client) LeaseRevoke(leaseId LeaseID) (*Header, error) {
	revoke, err := receiver.etcdCli.Revoke(todo, etcdV3.LeaseID(leaseId))
	if err != nil {
		_ = flog.Error(err)
		return nil, err
	}

	return newResponse(revoke.Header), nil
}

func (receiver *client) LeaseInfo(leaseId LeaseID) (*LeaseInfo, error) {
	leaseTimeToLiveResponse, err := receiver.etcdCli.TimeToLive(todo, etcdV3.LeaseID(leaseId))
	if err != nil {
		_ = flog.Error(err)
		return nil, err
	}

	return newLeaseInfo(leaseTimeToLiveResponse), nil
}

// UnLock 解锁
type UnLock func()

func (receiver *client) Lock(lockKey string, lockTTL int) (UnLock, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	session, err := concurrency.NewSession(receiver.etcdCli, concurrency.WithTTL(lockTTL))
	if err != nil {
		_ = flog.Error(err)
		return func() { cancelFunc() }, err
	}
	mutex := concurrency.NewMutex(session, lockKey)
	err = mutex.Lock(ctx)
	if err != nil {
		_ = flog.Error(err)
		return func() { cancelFunc() }, err
	}

	return func() {
		_ = mutex.Unlock(ctx)
		cancelFunc()
	}, nil
}
