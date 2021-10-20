package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var (
	config             clientv3.Config
	client             *clientv3.Client
	kv                 clientv3.KV
	err                error
	lease              clientv3.Lease
	leaseId            clientv3.LeaseID
	getResp            *clientv3.GetResponse
	leaseGrantResp     *clientv3.LeaseGrantResponse
	keepResp           *clientv3.LeaseKeepAliveResponse
	keepRespChan       <-chan *clientv3.LeaseKeepAliveResponse
	watchStartRevision int64
	watchRespChan      <-chan clientv3.WatchResponse
	watchResp          clientv3.WatchResponse
	event              *clientv3.Event
	watcher            clientv3.Watcher
)

func main() {
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
	//testFunc(client, rootContext)
	//testLease(client, rootContext)
	testWatch(client, rootContext)
	//testOp(client, rootContext)
}

//基本测试（获取值，设置值）
func testFunc(cli *clientv3.Client, rootContext context.Context) {
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
		fmt.Printf("empty for %s\n", kvs)
	}
	//设置值
	uuid := uuid.New().String()
	fmt.Printf("new value is :%s\r\n", uuid)
	ctx2, cancelFunc2 := context.WithTimeout(rootContext, time.Duration(2)*time.Second)
	_, err = kvc.Put(ctx2, "cc", uuid)
	if delRes, err := kvc.Delete(ctx2, "cc"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("delete %s for %d\n", "cc", delRes.Deleted)
	}
	cancelFunc2()
	if err != nil {
		fmt.Println(err)
	}
}

func testLease(client *clientv3.Client, rootContext context.Context) {
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

func testWatch(client *clientv3.Client, rootContext context.Context) {
	uuid := uuid.New().String()
	kv := clientv3.NewKV(client)

	// 模拟KV的变化
	go func() {
		for {
			_, err = kv.Put(context.TODO(), "bb", uuid)
			_, err = kv.Delete(context.TODO(), "bb")
			time.Sleep(1 * time.Second)
		}
	}()

	// 先GET到当前的值，并监听后续变化
	if getResp, err = kv.Get(context.TODO(), "bb"); err != nil {
		fmt.Println(err)
		return
	}

	// 现在key是存在的
	if len(getResp.Kvs) != 0 {
		fmt.Println("当前值:", string(getResp.Kvs[0].Value))
	}

	// 获得当前revision
	watchStartRevision = getResp.Header.Revision + 1
	// 创建一个watcher
	watcher = clientv3.NewWatcher(client)
	fmt.Println("从该版本向后监听:", watchStartRevision)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	time.AfterFunc(5*time.Second, func() {
		cancelFunc()
	})

	watchRespChan = watcher.Watch(ctx, "bb", clientv3.WithRev(watchStartRevision))
	// 处理kv变化事件
	for watchResp = range watchRespChan {
		for _, event = range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT:
				fmt.Println("修改为:", string(event.Kv.Value), "Revision:", event.Kv.CreateRevision, event.Kv.ModRevision)
			case mvccpb.DELETE:
				fmt.Println("删除了", "Revision:", event.Kv.ModRevision)
			}
		}
	}
}

func testOp(client *clientv3.Client, rootContext context.Context) {
	kv := clientv3.NewKV(client)
	uuid := uuid.New().String()

	putOp := clientv3.OpPut("aa", uuid)

	if opResp, err := kv.Do(context.TODO(), putOp); err != nil {
		panic(err)
	} else {
		fmt.Println("写入Revision:", opResp.Put().Header.Revision)
	}

	getOp := clientv3.OpGet("aa ")

	if opResp, err := kv.Do(context.TODO(), getOp); err != nil {
		panic(err)
	} else {
		fmt.Println("数据Revision:", opResp.Get().Kvs[0].ModRevision)
		fmt.Println("数据value:", string(opResp.Get().Kvs[0].Value))
	}

}

// Package clientv3 implements the official Go etcd client for v3.
//
// Create client using `clientv3.New`:
//
//	// expect dial time-out on ipv4 blackhole
//	_, err := clientv3.New(clientv3.Config{
//		Endpoints:   []string{"http://254.0.0.1:12345"},
//		DialTimeout: 2 * time.Second,
//	})
//
//	// etcd clientv3 >= v3.2.10, grpc/grpc-go >= v1.7.3
//	if err == context.DeadlineExceeded {
//		// handle errors
//	}
//
//	// etcd clientv3 <= v3.2.9, grpc/grpc-go <= v1.2.1
//	if err == grpc.ErrClientConnTimeout {
//		// handle errors
//	}
//
//	cli, err := clientv3.New(clientv3.Config{
//		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
//		DialTimeout: 5 * time.Second,
//	})
//	if err != nil {
//		// handle error!
//	}
//	defer cli.Close()
//
// Make sure to close the client after using it. If the client is not closed, the
// connection will have leaky goroutines.
//
// To specify a client request timeout, wrap the context with context.WithTimeout:
//
//	ctx, cancel := context.WithTimeout(context.Background(), timeout)
//	resp, err := kvc.Put(ctx, "sample_key", "sample_value")
//	cancel()
//	if err != nil {
//	    // handle error!
//	}
//	// use the response
//
// The Client has internal state (watchers and leases), so Clients should be reused instead of created as needed.
// Clients are safe for concurrent use by multiple goroutines.
//
// etcd client returns 3 types of errors:
//
//  1. context error: canceled or deadline exceeded.
//  2. gRPC status error: e.g. when clock drifts in server-side before client's context deadline exceeded.
//  3. gRPC error: see https://github.com/etcd-io/etcd/blob/master/etcdserver/api/v3rpc/rpctypes/error.go
//
// Here is the example code to handle client errors:
//
//	resp, err := kvc.Put(ctx, "", "")
//	if err != nil {
//		if err == context.Canceled {
//			// ctx is canceled by another routine
//		} else if err == context.DeadlineExceeded {
//			// ctx is attached with a deadline and it exceeded
//		} else if err == rpctypes.ErrEmptyKey {
//			// client-side error: key is not provided
//		} else if ev, ok := status.FromError(err); ok {
//			code := ev.Code()
//			if code == codes.DeadlineExceeded {
//				// server-side context might have timed-out first (due to clock skew)
//				// while original client-side context is not timed-out yet
//			}
//		} else {
//			// bad cluster endpoints, which are not etcd servers
//		}
//	}
//
//	go func() { cli.Close() }()
//	_, err := kvc.Get(ctx, "a")
//	if err != nil {
//		// with etcd clientv3 <= v3.3
//		if err == context.Canceled {
//			// grpc balancer calls 'Get' with an inflight client.Close
//		} else if err == grpc.ErrClientConnClosing { // <= gRCP v1.7.x
//			// grpc balancer calls 'Get' after client.Close.
//		}
//		// with etcd clientv3 >= v3.4
//		if clientv3.IsConnCanceled(err) {
//			// gRPC client connection is closed
//		}
//	}
//
// The grpc load balancer is registered statically and is shared across etcd clients.
// To enable detailed load balancer logging, set the ETCD_CLIENT_DEBUG environment
// variable.  E.g. "ETCD_CLIENT_DEBUG=1".
//
