package etcd

import (
	"github.com/farseer-go/fs/configure"
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
	etcdConfigs := configure.ParseConfigs[etcdConfig]("Etcd")
	for _, rabbitConfig := range etcdConfigs {
		if rabbitConfig.Server == "" {
			_ = flog.Error("Etcd配置缺少Server节点")
			continue
		}

	}
}

func (module Module) PostInitialize() {
}

func (module Module) Shutdown() {
}
