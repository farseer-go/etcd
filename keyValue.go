package etcd

import (
	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

type KeyValue struct {
	Header *Response
	// key 保存到etcd的KEY
	Key string
	// 创建KEY时的集群Revision（此后这个版本号不会再变）
	CreateRevision int64
	// 修改改这个 key 时的集群Revision（每次修改都会重新获取最新的集群Revision）
	ModRevision int64
	// 对KEY的任何修改，都会导致版本号增加（默认为1）
	Version int64
	// value
	Value string
	// 租约ID
	Lease int64
}

func newValue(kv *mvccpb.KeyValue, header *pb.ResponseHeader) *KeyValue {
	return &KeyValue{
		Header:         newResponse(header),
		Key:            string(kv.Key),
		CreateRevision: kv.CreateRevision,
		ModRevision:    kv.ModRevision,
		Version:        kv.Version,
		Value:          string(kv.Value),
		Lease:          kv.Lease,
	}
}

// Exists 是否有值
func (receiver *KeyValue) Exists() bool {
	return receiver.Version > 0 && receiver.CreateRevision > 0 && receiver.ModRevision > 0
}
