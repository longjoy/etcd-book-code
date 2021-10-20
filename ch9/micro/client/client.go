package main

import (
	"context"
	hello "github.com/keets2012/etcd-book-code/ch10/micro/srv/proto"
	"log"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/etcdv3"
)

func main() {
	reg := etcdv3.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{
			"106.15.233.99:2379",
		}
	})

	service := micro.NewService(
		micro.Registry(reg),
	)
	service.Init()

	sayClient := hello.NewGreeterService("hello.srv.say", service.Client())
	param := &hello.HelloRequest{
		From: "client",
		To:   "server",
		Msg:  "hello aoho",
	}

	rsp, err := sayClient.Hello(context.Background(), param)
	if err != nil {
		panic(err)
	}

	log.Println(rsp)
}
