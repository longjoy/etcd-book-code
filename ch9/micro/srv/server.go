package main

import (
	"context"
	hello "github.com/keets2012/etcd-book-code/ch10/micro/srv/proto"
	"log"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/etcdv3"
)

type Greet struct{}

func (s *Greet) Hello(ctx context.Context, req *hello.HelloRequest, rsp *hello.HelloResponse) error {
	log.Printf("received req %#v \n", req)
	rsp.From = "server"
	rsp.To = "client"
	rsp.Msg = "ok"
	return nil
}

func main() {
	reg := etcdv3.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{"106.15.233.99:2379",
		}
	})

	service := micro.NewService(
		micro.Name("hello.srv.say"),
		micro.Registry(reg),
	)
	service.Init()

	hello.RegisterGreeterHandler(service.Server(), new(Greet))

	if err := service.Run(); err != nil {
		panic(err)
	}
}
