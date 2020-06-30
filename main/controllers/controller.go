package controllers

import (
	"../tools"
	"container/list"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type Worker struct {
	Ws  *websocket.Conn
	Id  int
	Cmd map[string]interface{}
}
type Controller struct {
	beego.Controller
	Identify int
	Events   chan string
	Clients  *list.List
	Ws       *websocket.Upgrader
	Mutex    *sync.Mutex
}

func (c *Controller) Index() {
	c.TplName = "index.html"
}

func (c *Controller) GetDownloadFile() {
	file := c.GetString("file", "")
	kind, _ := c.GetInt("kind", 1)
	if file == "" {
		var e = &tools.ErrorRes{
			Code:    404,
			Message: "File Not Found",
		}
		c.Data["json"] = e
		c.Ctx.Output.SetStatus(404)
		c.ServeJSON()
		return
	}
	if kind == 1 {
		c.Ctx.Output.Download("./File/Douyin/"+file, file)
	}
	if kind == 2 {
		c.Ctx.Output.Download("./File/Instagram/"+file, file)
	}
	if kind == 3 {
		c.Ctx.Output.Download("./File/Facebook/"+file, file)
	}

}

func (c *Controller) WsConnect() {
	ws, err := c.Ws.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(c.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		beego.Error("Cannot setup WebSocket connection:", err)
		return
	}
	defer ws.Close()
	for {
		var req tools.Req
		//var rq interface{}
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&req)
		if err != nil {
			fmt.Printf("error: %v", err)
			c.StopRun()
		}
		// Send the newly received message to the broadcast channel
		if req.Kind == 0 {
			//var path = tools.MainWorkFlow(&req, ws, c.Mutex)
			//c.Events <- path
			var file = tools.DownloadDouyin(req.Url)
			if file == ""{
				_ = ws.WriteJSON(&tools.Info{
					AuthorName: req.Username,
					Id:         "",
					Follow:     0,
					Region:     "",
					Sign:       "",
					State:      0,
					Result:     "",
					Total:      100,
					Progress:   100,
				})
				c.StopRun()
			}
			_ = ws.WriteJSON(&tools.Info{
				AuthorName: req.Url,
				Id:         "",
				Follow:     0,
				Region:     "",
				Sign:       "",
				State:      2,
				Result:     "http://65.52.184.198/download?kind=1&file=" + file,
				Total:      100,
				Progress:   100,
			})
		}
		if req.Kind == 1 {
			fmt.Println(req.Cookies)
			var file, err = tools.DownloadFileIG(req.Username, req.Cookies)
			if err != nil {
				_ = ws.WriteJSON(&tools.Info{
					AuthorName: req.Username,
					Id:         "",
					Follow:     0,
					Region:     "",
					Sign:       "",
					State:      0,
					Result:     "",
					Total:      100,
					Progress:   100,
				})
				c.StopRun()
			}

			_ = ws.WriteJSON(&tools.Info{
				AuthorName: req.Username,
				Id:         "",
				Follow:     0,
				Region:     "",
				Sign:       "",
				State:      2,
				Result:     "http://65.52.184.198/download?kind=2&file=" + file,
				Total:      100,
				Progress:   100,
			})
		}
		if req.Kind == 2 {
			fmt.Println(req.Cookies)
			var file = tools.DownloadFileFb(req.Url)
			_ = ws.WriteJSON(&tools.Info{
				AuthorName: req.Url,
				Id:         "",
				Follow:     0,
				Region:     "",
				Sign:       "",
				State:      2,
				Result:     "http://65.52.184.198/download?kind=3&file=" + file,
				Total:      100,
				Progress:   100,
			})
		}
	}

}

func handleDownload(ws *websocket.Conn, data map[string]interface{}) {
	if data["event"] == "tiktok" {

	}
}
