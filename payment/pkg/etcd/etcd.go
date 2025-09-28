package etcd

import (
	"bytedancemall/payment/config"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var etcdClient *clientv3.Client

func init() {
	var err error
	if etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   config.Cfg.Etcd.Host,
		DialTimeout: 5 * 1e9,
	}); err != nil {
		panic(err)
	}
}

func GetEtcdCli() *clientv3.Client {
	return etcdClient
}
