package main

import (
	"apiserver_demos/demo01/router"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

func main() {
	// Create the Gin engine.
	g := gin.New()

	//type HandlerFunc func(*Context) 定义函数别名HandlerFunc, 函数参数为*Context，无返回值
	//middlewares为函数数组，函数类型为HandlerFunc，{}函数数组的初始化值，这里为空
	middlewares := []gin.HandlerFunc{}

	log.Printf("middlewares size %d", len(middlewares))

	// Routes.
	router.Load(
		// Cores.
		g,

		// Middlwares.省略号展开数组
		middlewares...,
	)

	// Ping the server to make sure the router is working.
	// 开起协程
	go func() {
		if err := pingServer(); err != nil {
			//终止主进程是通过log.Fatal
			log.Fatal("The router has no response, or it might took too long to start up.", err)
		}
		log.Print("The router has been deployed successfully.")
	}()

	log.Printf("Start to listening the incoming requests on http address: %s", ":8080")
	log.Printf(http.ListenAndServe(":8080", g).Error())
}

// pingServer pings the http server to make sure the router is working.
func pingServer() error {
	for i := 0; i < 2; i++ {
		// Ping the server by sending a GET request to `/health`.
		resp, err := http.Get("http://127.0.0.1:8080" + "/sd/health")
		if err == nil && resp.StatusCode == 200 {
			return nil
		}

		// Sleep for a second to continue the next ping.
		log.Print("Waiting for the router, retry in 1 second.")
		time.Sleep(time.Second)
	}
	return errors.New("Cannot connect to the router.")
}
