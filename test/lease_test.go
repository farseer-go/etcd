package test

import (
	"github.com/farseer-go/etcd"
	"github.com/farseer-go/fs"
	"github.com/farseer-go/fs/container"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLease(t *testing.T) {
	fs.Initialize[etcd.Module]("test etcd")
	client := container.Resolve[etcd.IClient]("default")

	_, _ = client.PutJson("/test/lease3", []int{3})

	leaseID, _ := client.LeaseGrant(1, "/test/lease3")

	_, _ = client.PutLease("/test/lease1", "1", leaseID)
	_, _ = client.PutJsonLease("/test/lease2", []int{2}, leaseID)

	info, _ := client.LeaseInfo(leaseID)
	assert.Equal(t, int64(2), info.GrantedTTL)

	time.Sleep(2500 * time.Millisecond)
	assert.False(t, client.Exists("/test/lease1"))
	assert.False(t, client.Exists("/test/lease2"))
	assert.False(t, client.Exists("/test/lease3"))
}
