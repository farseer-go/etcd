package test

import (
	"context"
	"github.com/farseer-go/etcd"
	"github.com/farseer-go/fs"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/flog"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	fs.Initialize[etcd.Module]("test etcd")
	etcd.Module{}.Shutdown()
	client := container.Resolve[etcd.IClient]("default")
	assert.Equal(t, "", client.Original().Username)

	watchResult := make(map[string]string)
	client.Watch(context.TODO(), "/test/b1", func(event etcd.WatchEvent) {
		watchResult[event.Kv.Key] = event.Kv.Value
		event.IsModify()
	})

	client.WatchPrefixKey(context.TODO(), "/test/", func(event etcd.WatchEvent) {
		watchResult[event.Kv.Key] = event.Kv.Value
		event.IsCreate()
	})

	putRsp, err := client.Put("/test/a1", "1")
	assert.NoError(t, err)
	assert.Less(t, int64(0), putRsp.Revision)
	flog.Info(putRsp.String())

	putRsp, err = client.Put("/test/a2", "2")
	assert.NoError(t, err)

	putRsp, err = client.Put("/test/b1", "3")
	assert.NoError(t, err)

	putRsp, err = client.PutJson("/test/a1/b1", []int{4})
	assert.NoError(t, err)

	result, err := client.Get("/test/a1")
	assert.Equal(t, "1", result.Value)
	assert.Equal(t, putRsp.Revision, result.Header.Revision)

	result, err = client.Get("/test/a2")
	assert.Equal(t, "2", result.Value)
	assert.Equal(t, putRsp.Revision, result.Header.Revision)

	result, err = client.Get("/test/a1/b1")
	assert.Equal(t, "[4]", result.Value)
	assert.Equal(t, putRsp.Revision, result.Header.Revision)

	results, err := client.GetPrefixKey("/test")
	assert.Equal(t, "1", results["/test/a1"].Value)
	assert.Equal(t, "2", results["/test/a2"].Value)
	assert.Equal(t, "[4]", results["/test/a1/b1"].Value)
	assert.Equal(t, "3", results["/test/b1"].Value)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "1", watchResult["/test/a1"])
	assert.Equal(t, "2", watchResult["/test/a2"])
	assert.Equal(t, "[4]", watchResult["/test/a1/b1"])
	assert.Equal(t, "3", watchResult["/test/b1"])

	_, _ = client.Delete("/test/a1/b1")
	result, err = client.Get("/test/a1/b1")
	assert.False(t, result.Exists())

	_, _ = client.DeletePrefixKey("/te")

	result, err = client.Get("/test/a1")
	assert.False(t, result.Exists())

	result, err = client.Get("/test/a2")
	assert.False(t, result.Exists())

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "", watchResult["/test/a1"])
	assert.Equal(t, "", watchResult["/test/a2"])
	assert.Equal(t, "", watchResult["/test/a1/b1"])
	assert.Equal(t, "", watchResult["/test/b1"])

	client.Close()
}
