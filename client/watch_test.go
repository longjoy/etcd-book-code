package client

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"testing"
	"time"
)

var (
	watchStartRevision int64
	watchRespChan      <-chan clientv3.WatchResponse
	watchResp          clientv3.WatchResponse
	event              *clientv3.Event
	watcher            clientv3.Watcher
)

func TestWatch(t *testing.T) {
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

	uuid := uuid.New().String()
	kv := clientv3.NewKV(client)
	//var err error = nil
	// 模拟KV的变化
	go func() {
		for {
			_, err = kv.Put(rootContext, "bb", uuid)
			_, err = kv.Delete(rootContext, "bb")
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
