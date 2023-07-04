package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/igm/sockjs-go/v3/sockjs"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/caoyingjunz/kube-webshell/app"
)

func main() {
	r := gin.Default()

	// 静态文件和 html 文件引入
	r.Static("./static", "./static")
	r.LoadHTMLGlob("templates/*")

	r.GET("", func(c *gin.Context) {
		c.Request.URL.Path = "/index"
		r.HandleContext(c)
	})
	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.GET("/webshell", func(c *gin.Context) {
		var query struct {
			Namespace string `form:"namespace"`
			Pod       string `form:"pod"`
			Container string `form:"container"`
		}
		if err := c.ShouldBindQuery(&query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.HTML(http.StatusOK, "webshell.html", gin.H{
			"namespace": query.Namespace,
			"pod":       query.Pod,
			"container": query.Container,
		})
	})

	r.GET("/webshell/ws/*info", func(c *gin.Context) {
		var query struct {
			Namespace string `form:"namespace"`
			Pod       string `form:"pod"`
			Container string `form:"container"`
		}
		if err := c.ShouldBindQuery(&query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		sockjs.NewHandler("/webshell/ws", sockjs.DefaultOptions, func(session sockjs.Session) {
			if err := app.WebShellHandler(&app.WebShell{
				Conn:      session,
				SizeChan:  make(chan *remotecommand.TerminalSize),
				Namespace: query.Namespace,
				Pod:       query.Pod,
				Container: query.Container,
			}, "/bin/bash"); err != nil {
				if err = app.WebShellHandler(&app.WebShell{
					Conn:      session,
					SizeChan:  make(chan *remotecommand.TerminalSize),
					Namespace: query.Namespace,
					Pod:       query.Pod,
					Container: query.Container,
				}, "/bin/sh"); err != nil {
					fmt.Print(err)
				}
			}
		}).ServeHTTP(c.Writer, c.Request)
	})

	_ = r.Run(":8080")
}
