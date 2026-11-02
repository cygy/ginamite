// Command helloworld-api is the smallest API built with ginamite: a single route, GET /v1/hello, which answers {"hello":"world"}.
package main

import (
	"log"
	"net/http"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/api/route"

	"github.com/gin-gonic/gin"
)

func main() {
	server := ginamite.NewServer()

	server.API.AddCustomRoutes(func(group *gin.RouterGroup, _ route.Middlewares, _ bool, _ func(bool, []string) gin.HandlerFunc) {
		group.GET("/hello", hello)
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

// hello answers {"hello":"world"}.
func hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"hello": "world"})
}
