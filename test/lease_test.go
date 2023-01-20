package test

import (
	"github.com/farseer-go/etcd"
	"github.com/farseer-go/fs/container"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLease(t *testing.T) {
	client := container.Resolve[etcd.IClient]("default")
	leaseID, _ := client.LeaseGrant(1)
	_, _ = client.PutLease("/test/lease1", "1", leaseID)
	_, _ = client.PutJsonLease("/test/lease2", []int{1}, leaseID)

	assert.True(t, client.Exists("/test/lease1"))
	assert.True(t, client.Exists("/test/lease2"))
	time.Sleep(2500 * time.Millisecond)
	assert.False(t, client.Exists("/test/lease1"))
	assert.False(t, client.Exists("/test/lease2"))
}
