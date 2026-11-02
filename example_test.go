package ginamite_test

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/plugins/kafka"
	"github.com/cygy/ginamite/plugins/mongodb"

	"github.com/gin-gonic/gin"
)

// An application registers its plugins and its routes, then runs the server the command line arguments choose, until
// SIGINT or SIGTERM.
func Example() {
	server := ginamite.NewServer()
	server.AddPlugin(mongodb.NewPlugin())
	server.AddPlugin(kafka.NewPlugin())

	server.API.AddCustomRoutes(func(group *gin.RouterGroup, _ route.Middlewares, _ bool, _ func(bool, []string) gin.HandlerFunc) {
		group.GET("/hello", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"hello": "world"}) })
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
