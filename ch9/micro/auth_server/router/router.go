package router

import (
	"context"
	"github.com/gin-gonic/gin"
	hello "github.com/keets2012/etcd-book-code/ch10/micro/srv/proto"
	"github.com/micro/go-micro"
	"log"
)

func InitRouters(service micro.Service) *gin.Engine {
	ginRouter := gin.Default()
	ginRouter.POST("/orders/", func(context *gin.Context) {
		res := callUser(service)
		context.String(200, res.String())
	})

	return ginRouter
}

func callUser(service micro.Service) *hello.SayResponse {

	sayClient := hello.NewSayService("user.server", service.Client())
	param := &hello.SayRequest{
		From: "client",
		To:   "server",
		Msg:  "hello aoho",
	}

	rsp, err := sayClient.Hello(context.Background(), param)
	if err != nil {
		panic(err)
	}

	log.Println(rsp)
	return rsp
}
