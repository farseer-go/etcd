package test

import (
	"github.com/farseer-go/etcd"
	"github.com/farseer-go/fs"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/flog"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	fs.Initialize[etcd.Module]("test etcd")
	client := container.Resolve[etcd.IClient]("default")

	result := 0
	unLock1, _ := client.Lock("/lock/1", 1)
	flog.Info("上锁：unLock1")
	go func() {
		time.Sleep(3000 * time.Millisecond)
		flog.Info("解锁：unLock1")
		unLock1()
		result++
	}()
	result++
	unLock2, _ := client.Lock("/lock/1", 3)
	flog.Info("上锁：unLock2")
	assert.Equal(t, 2, result)
	unLock2()
	flog.Info("解锁：unLock2")
}
