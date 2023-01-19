package etcd

import (
	"context"
	"encoding/json"
	"github.com/farseer-go/fs/flog"
	etcdV3 "go.etcd.io/etcd/client/v3"
	"strings"
	"time"
)

var todo = context.TODO()

type etcdClient = etcdV3.Client

type client struct {
	etcdCli *etcdClient
}

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

func (receiver *client) PutJson(key string, data any) (*Header, error) {
	jsonValue, _ := json.Marshal(data)
	rsp, err := receiver.etcdCli.Put(todo, key, string(jsonValue))
	if err != nil {
		_ = flog.Error(err)
		return nil, err
	}
	return newResponse(rsp.Header), err
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
