package etcd

// WatchEvent 监听事件
type WatchEvent struct {
	Type string    // PUT 或者 DELETE
	Kv   *KeyValue // 最新的KV信息
}

// IsCreate 是否为创建事件
func (e *WatchEvent) IsCreate() bool {
	return e.Type == "PUT" && e.Kv.CreateRevision == e.Kv.ModRevision
}

// IsModify 是否为修改事件
func (e *WatchEvent) IsModify() bool {
	return e.Type == "DELETE" && e.Kv.CreateRevision != e.Kv.ModRevision
}
