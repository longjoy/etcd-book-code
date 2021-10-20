package client

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"testing"
	"time"
)

func TestOp(t *testing.T) {

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"106.15.233.99:2379"},
		DialTimeout: 2 * time.Second,
	})
	// etcd clientv3 >= v3.2.10, grpc/grpc-go >= v1.7.3
	if client == nil || err == context.DeadlineExceeded {
		// handle errors
		fmt.Println(err)
		panic("invalid connection!")
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	uuid := uuid.New().String()

	putOp := clientv3.OpPut("aa", uuid)

	if opResp, err := kv.Do(context.TODO(), putOp); err != nil {
		panic(err)
	} else {
		fmt.Println("写入Revision:", opResp.Put().Header.Revision)
	}

	getOp := clientv3.OpGet("aa")

	if opResp, err := kv.Do(context.TODO(), getOp); err != nil {
		panic(err)
	} else {
		fmt.Println("数据Revision:", opResp.Get().Kvs[0].ModRevision)
		fmt.Println("数据value:", string(opResp.Get().Kvs[0].Value))
	}
}
