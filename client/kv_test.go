package client

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"testing"
	"time"
)

func TestKV(t *testing.T) {
	rootContext := context.Background()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"106.15.233.99:2379"},
		DialTimeout: 2 * time.Second,
	})
	// etcd clientv3 >= v3.2.10, grpc/grpc-go >= v1.7.3
	if cli == nil || err == context.DeadlineExceeded {
		// handle errors
		fmt.Println(err)
		panic("invalid connection!")
	}
	defer cli.Close()

	kvc := clientv3.NewKV(cli)
	//获取值
	ctx, cancelFunc := context.WithTimeout(rootContext, time.Duration(2)*time.Second)
	response, err := kvc.Get(ctx, "cc")
	cancelFunc()
	if err != nil {
		fmt.Println(err)
	}
	kvs := response.Kvs
	if len(kvs) > 0 {
		fmt.Printf("last value is :%s\r\n", string(kvs[0].Value))
	} else {
		fmt.Printf("empty key for %s\n", "cc")
	}
	//设置值
	uuid := uuid.New().String()
	fmt.Printf("new value is :%s\r\n", uuid)
	ctx2, cancelFunc2 := context.WithTimeout(rootContext, time.Duration(2)*time.Second)
	_, err = kvc.Put(ctx2, "cc", uuid)
	if delRes, err := kvc.Delete(ctx2, "cc"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("delete %s for %t\n", "cc", delRes.Deleted > 0)
	}
	cancelFunc2()
	if err != nil {
		fmt.Println(err)
	}
}
