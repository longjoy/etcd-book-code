package client

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"testing"
	"time"
)

var (
	leaseId        clientv3.LeaseID
	getResp        *clientv3.GetResponse
	leaseGrantResp *clientv3.LeaseGrantResponse
	kv             clientv3.KV
	keepResp       *clientv3.LeaseKeepAliveResponse
	keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
)

func TestLease(t *testing.T) {
	rootContext := context.Background()

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

	// 申请一个租约
	lease := clientv3.NewLease(client)
	if leaseGrantResp, err = lease.Grant(context.TODO(), 10); err != nil {
		fmt.Println(err)
		return
	}
	leaseId = leaseGrantResp.ID
	// 申请一个租约
	lease = clientv3.NewLease(client)
	keepLease(lease, int64(leaseId))

	// 获得kv API子集
	kv = clientv3.NewKV(client)
	uuid := uuid.New().String()
	ctx, cancelFunc := context.WithTimeout(rootContext, time.Duration(2)*time.Second)
	if _, err = kv.Put(ctx, "dd", uuid, clientv3.WithLease(leaseId)); err != nil {
		fmt.Println(err)
		return
	}
	cancelFunc()
	for {
		ctx2, cancelFunc2 := context.WithTimeout(rootContext, time.Duration(2)*time.Second)
		if getResp, err = kv.Get(ctx2, "dd"); err != nil {
			fmt.Println(err)
			return
		}
		cancelFunc2()
		if getResp.Count == 0 {
			fmt.Println("kv过期了")
			break
		}
		fmt.Println("还没过期:", getResp.Kvs)
		time.Sleep(2 * time.Second)
	}

}

func keepLease(lease clientv3.Lease, leaseId int64) {
	var err error = nil
	if keepRespChan, err = lease.KeepAlive(context.TODO(), clientv3.LeaseID(leaseId)); err != nil {
		fmt.Println("error is : ", err)
		return
	}
	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepRespChan == nil {
					fmt.Println("租约已经失效了")
					goto END
				} else { // 每秒会续租一次, 所以就会受到一次应答
					fmt.Println("收到自动续租应答:", keepResp.ID)
				}
			}
		}
	END:
	}()
}
