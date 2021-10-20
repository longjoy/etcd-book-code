package client

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"testing"
	"time"
)

func TestEtcdClientInit(t *testing.T) {

	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
	)
	// 客户端配置
	config = clientv3.Config{
		Endpoints:   []string{"106.15.233.99:2379"},
		DialTimeout: 5 * time.Second,
	}
	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(client.Cluster.MemberList(context.TODO()))
	}
	client.Close()
}
