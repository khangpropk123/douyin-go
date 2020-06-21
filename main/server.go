package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"sync"
	"time"
)
import "./controllers"

func main() {
	var controller = &controllers.Controller{
		Identify: 0,
		Events:   make(chan string),
		Ws: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Mutex: &sync.Mutex{},
	}
	go func() {
		for {
			select {
			case data := <-controller.Events:
				fmt.Println(data)
				go func() {
					time.Sleep(time.Second * 3600)
					err := os.Remove(data)
					fmt.Println(err)
				}()
			}
		}
	}()
	beego.Router("/tiktok", controller, "get:WsConnect")
	beego.Router("/download", controller, "get:GetDownloadFile")
	beego.Router("/", controller, "get:Index")
	beego.SetStaticPath("/", "views")
	beego.Run()
}
