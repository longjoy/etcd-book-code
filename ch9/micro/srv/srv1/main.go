package main

import (
	"context"
	demo "github.com/keets2012/etcd-book-code/ch10/micro/srv/proto/demo"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/etcdv3"
	_ "github.com/micro/go-plugins/registry/etcdv3"
	"log"
	"math/rand"
	"time"
)

type Say struct{}

func (s *Say) Hello(ctx context.Context, req *demo.Request, rsp *demo.Response) error {
	log.Print("Received Say.Hello request")
	rsp.Msg = "Hello " + req.Name + " srv one. rand:" + string(rand.Intn(256))
	return nil
}

func main() {
	reg := etcdv3.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{"106.15.233.99:2379"}
	})
	service := micro.NewService(
		micro.Name("go.micro.srv.greeter"),
		micro.Registry(reg),
		micro.RegisterTTL(time.Second*30),      // 这是设置注册到etcd那个key的过期时间
		micro.RegisterInterval(time.Second*10), // 这是服务去etcd报告自己还活着的周期
		// 服务在下线的时候会去 etcd 卸载自己，但是遇到进程杀掉，网络不通情况就让这个ttl让服务下线
	)

	// optionally setup command line usage
	service.Init()

	// Register Handlers
	demo.RegisterSayHandler(service.Server(), new(Say))

	// Run server
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
