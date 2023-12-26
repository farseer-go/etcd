package test

import (
	"github.com/farseer-go/etcd"
	"github.com/farseer-go/fs"
	"github.com/farseer-go/fs/configure"
)

func init() {
	// 设置配置默认值，模拟配置文件
	configure.SetDefault("Etcd.default", "Server=127.0.0.1:2379|127.0.0.1:2379,DialTimeout=5000")
	fs.Initialize[etcd.Module]("test etcd")
}
