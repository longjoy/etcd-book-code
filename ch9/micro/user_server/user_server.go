package main

import (
	"github.com/keets2012/etcd-book-code/ch10/micro/user_server/router"
	"github.com/micro/go-micro/registry"      //
	"github.com/micro/go-micro/registry/etcd" //
	"github.com/micro/go-micro/web"           //
)

var etcdReg registry.Registry

func init() {
	//新建一个consul注册的地址，也就是我们consul服务启动的机器ip+端口
	etcdReg = etcd.NewRegistry(
		registry.Addrs("106.15.233.99:2379"),
	)
}

func main() {
	//初始化路由
	ginRouter := router.InitRouters()

	//注册服务
	microService := web.NewService(
		web.Name("user.server"),
		//web.RegisterTTL(time.Second*30),//设置注册服务的过期时间
		//web.RegisterInterval(time.Second*20),//设置间隔多久再次注册服务
		web.Address(":8001"),
		web.Handler(ginRouter),
		web.Registry(etcdReg),
	)

	microService.Run()
}
