package main

import (
	"application/pkg/pubsub"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type service struct {
	count uint
	ps    pubsub.PubSub
	sync.Mutex
}

func main() {
	svc := service{
		count: 0,
		ps:    pubsub.New(),
	}

	app := gin.New()
	app.GET("/counter", svc.counter)
	app.GET("/live", svc.live)
	app.Run("0.0.0.0:8000")
}

func (svc *service) counter(ctx *gin.Context) {
	svc.Lock()
	svc.count++
	svc.Unlock()
	svc.ps.Publish("counter", svc.count)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "counter incremented",
		"data": gin.H{
			"count": svc.count,
		},
	})
}

func (svc *service) live(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}
	for count := range svc.ps.Subscribe("counter") {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"message": "get live counter succeed", "data": {"count": %d}}`, count.(uint)))); err != nil {
			return
		}
	}
}
