package test

import (
	"context"
	"github.com/farseer-go/etcd"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/flog"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLease(t *testing.T) {
	client := container.Resolve[etcd.IClient]("default")

	_, _ = client.PutJson("/test/lease3", []int{3})

	leaseID, _ := client.LeaseGrant(1, "/test/lease3")

	_, _ = client.PutLease("/test/lease1", "1", leaseID)
	_, _ = client.PutJsonLease("/test/lease2", []int{2}, leaseID)

	info, _ := client.LeaseInfo(leaseID)
	assert.Equal(t, int64(2), info.GrantedTTL)

	assert.True(t, client.Exists("/test/lease1"))
	assert.True(t, client.Exists("/test/lease2"))
	assert.True(t, client.Exists("/test/lease3"))
	time.Sleep(2500 * time.Millisecond)
	assert.False(t, client.Exists("/test/lease1"))
	assert.False(t, client.Exists("/test/lease2"))
	assert.False(t, client.Exists("/test/lease3"))

	// 持续续约
	flog.Info("持续续约")
	leaseID, _ = client.LeaseGrant(1)
	_, _ = client.PutLease("/test/lease4", "1", leaseID)
	ctx, cancelFunc := context.WithCancel(context.Background())
	_ = client.LeaseKeepAlive(ctx, leaseID)
	time.Sleep(2500 * time.Millisecond)
	assert.True(t, client.Exists("/test/lease4"))
	_, _ = client.LeaseRevoke(leaseID)
	assert.False(t, client.Exists("/test/lease4"))
	cancelFunc()

	// 续约一次
	leaseID, _ = client.LeaseGrant(1)
	_, _ = client.PutLease("/test/lease5", "1", leaseID)
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)
		_ = client.LeaseKeepAliveOnce(leaseID)
		flog.Info("续约一次")
	}
	assert.True(t, client.Exists("/test/lease5"))
	time.Sleep(2500 * time.Millisecond)
	assert.False(t, client.Exists("/test/lease5"))
}
