package etcd

import (
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/fs/modules"
)

type Module struct {
}

func (module Module) DependsModule() []modules.FarseerModule {
	return nil
}

func (module Module) PreInitialize() {
}

func (module Module) Initialize() {
	etcdConfigs := configure.GetSubNodes("Etcd")
	for name, configString := range etcdConfigs {
		config := configure.ParseString[etcdConfig](configString.(string))
		if config.Server == "" {
			_ = flog.Error("Etcd配置缺少Server节点")
			continue
		}

		// 注册实例
		container.RegisterTransient(func() IClient {
			return newClient(config)
		}, name)
	}
}

func (module Module) PostInitialize() {
}

func (module Module) Shutdown() {
}
