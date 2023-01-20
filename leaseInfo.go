package etcd

import (
	etcdV3 "go.etcd.io/etcd/client/v3"
)

type LeaseInfo struct {
	Header     *Header
	ID         LeaseID  // 租约ID
	TTL        int64    // TTL是租约的剩余TTL（秒）；租约将在TTL+1秒下过期。过期的租约将返回-1。
	GrantedTTL int64    // 创建租约时设定的时间（单位s）
	Keys       []string // 租约关联的KEY
}

func newLeaseInfo(timeToLiveResponse *etcdV3.LeaseTimeToLiveResponse) *LeaseInfo {
	var keys []string
	for _, bytes := range timeToLiveResponse.Keys {
		keys = append(keys, string(bytes))
	}

	return &LeaseInfo{
		Header:     newResponse(timeToLiveResponse.ResponseHeader),
		ID:         LeaseID(timeToLiveResponse.ID),
		TTL:        timeToLiveResponse.TTL,
		GrantedTTL: timeToLiveResponse.GrantedTTL,
		Keys:       keys,
	}
}
