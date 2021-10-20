package main

import (
	"fmt"
	"github.com/keets2012/etcd-book-code/ch10/micro/auth_server/router"
	"github.com/micro/go-micro/client/selector"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/registry/etcd"
	"github.com/micro/go-micro/web"
	"net/http"
	"time"
)

var etcdReg registry.Registry

func init() {
	//新建一个 consul 注册的地址，也就是我们consul服务启动的机器ip+端口
	etcdReg = etcd.NewRegistry(
		registry.Addrs("106.15.233.99:2379"),
	)
}


func main() {
	//初始化路由
	ginRouter := router.InitRouters()

	//注册服务
	microService := web.NewService(
		web.Name("auth.server"),
		//web.RegisterTTL(time.Second*30),//设置注册服务的过期时间
		//web.RegisterInterval(time.Second*20),//设置间隔多久再次注册服务
		web.Address(":8002"),
		web.Handler(ginRouter),
		web.Registry(etcdReg),
	)

	//获取服务地址
	hostAddress := GetServiceAddr("user.server")
	if len(hostAddress) <= 0 {
		fmt.Println("hostAddress is null")
	} else {
		url := "http://" + hostAddress + "/users"
		response, _ := http.Get(url)
		fmt.Println(response.StatusCode)
	}

	microService.Run()
}

func GetServiceAddr(serviceName string) (address string) {
	var retryCount int
	for {
		servers, err := etcdReg.GetService(serviceName)
		if err != nil {
			fmt.Println(err.Error())
		}
		var services []*registry.Service
		for _, value := range servers {
			fmt.Println(value.Name, ":", value.Version)
			services = append(services, value)
		}
		next := selector.RoundRobin(services)
		if node, err := next(); err == nil {
			address = node.Address
		}
		if len(address) > 0 {
			return
		}
		//重试次数++
		retryCount++
		time.Sleep(time.Second * 1)
		//重试5次为获取返回空
		if retryCount >= 5 {
			return
		}
	}
}
