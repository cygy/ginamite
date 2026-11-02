// Command helloworld-web is the smallest web server built with ginamite: a single page, /en-us/hello, rendered from an
// HTML template. Nothing else runs: no API, no worker.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/route"

	"github.com/gin-gonic/gin"
)

func main() {
	server := ginamite.NewServer()

	server.Web.AddCustomRoutes(configureRoutes)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

// configureRoutes adds, for each locale, the route of the page "index" of the routes file.
func configureRoutes(g *gin.RouterGroup, _ string, _ route.Middlewares, getRelativePath func(string) (string, bool, gin.HandlerFunc, bool)) {
	if path, _, savePageKey, ok := getRelativePath("index"); ok {
		g.GET(path, savePageKey, hello)
	}
}

// hello renders the page "index".
func hello(c *gin.Context) {
	page := html.Of(c).PageByKey(c, "index")
	page.Content.Title = page.Title
	page.RenderOk(c)
}
