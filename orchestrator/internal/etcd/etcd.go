package etcd

type Etcd struct {
	path string
}

func NewEtcd() *Etcd {
	return &Etcd{}
}

func (r *Etcd) LoadNodeSettings() {
	return
}
