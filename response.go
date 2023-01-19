package etcd

import (
	"fmt"
	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
)

type Response struct {
	// 集群ID
	ClusterId uint64
	// 处理本次请求的节点ID
	MemberId uint64
	// 集群的版本号（该集群的任何KEY有变化时，版本号都会增加）
	Revision int64
	// raft_term is the raft term when the request was applied.
	RaftTerm uint64
}

func newResponse(header *pb.ResponseHeader) *Response {
	return &Response{
		ClusterId: header.ClusterId,
		MemberId:  header.MemberId,
		Revision:  header.Revision,
		RaftTerm:  header.RaftTerm,
	}
}

func (receiver *Response) String() string {
	return fmt.Sprintf("cluster_id:%d member_id:%d revision:%d raft_term:%d ", receiver.ClusterId, receiver.MemberId, receiver.Revision, receiver.RaftTerm)
}
