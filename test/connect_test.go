package test

import (
	"github.com/farseer-go/etcd"
	"github.com/farseer-go/fs"
	"github.com/farseer-go/fs/container"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnect(t *testing.T) {
	fs.Initialize[etcd.Module]("test etcd")
	etcd.Module{}.Shutdown()
	client := container.Resolve[etcd.IClient]("default")
	putRsp, err := client.Put("/test/a1", "1")
	assert.NoError(t, err)
	assert.Less(t, int64(0), putRsp.Revision)

	putRsp, err = client.Put("/test/a2", "2")
	assert.NoError(t, err)

	putRsp, err = client.Put("/test/a1/b1", "3")
	assert.NoError(t, err)

	result, err := client.Get("/test/a1")
	assert.Equal(t, "1", result.Value)
	assert.Equal(t, putRsp.Revision, result.Header.Revision)

	result, err = client.Get("/test/a2")
	assert.Equal(t, "2", result.Value)
	assert.Equal(t, putRsp.Revision, result.Header.Revision)

	result, err = client.Get("/test/a1/b1")
	assert.Equal(t, "3", result.Value)
	assert.Equal(t, putRsp.Revision, result.Header.Revision)

	results, err := client.GetPrefixKey("/test")
	assert.Equal(t, "1", results["/test/a1"].Value)
	assert.Equal(t, "2", results["/test/a2"].Value)
	assert.Equal(t, "3", results["/test/a1/b1"].Value)

	_, _ = client.Delete("/test/a1/b1")
	result, err = client.Get("/test/a1/b1")
	assert.False(t, result.Exists())

	_, _ = client.DeletePrefixKey("/te")

	result, err = client.Get("/test/a1")
	assert.False(t, result.Exists())

	result, err = client.Get("/test/a2")
	assert.False(t, result.Exists())
}
