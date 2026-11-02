# Custom routes

The routes of the application are added before `Run`, by a function called when the server sets up its routes.

**API.** The function receives the group of the version of the API (`/v1`), the built-in middlewares, whether the
environment is the production, and a function returning the middleware which checks the rights of the user:

```go
server.API.AddCustomRoutes(func(group *gin.RouterGroup, handlers route.Middlewares, isProd bool, rights func(admin bool, abilities []string) gin.HandlerFunc) {
	group.GET("/public", public)
	group.GET("/me", handlers.MustBeAuthenticated, handlers.MustAuthenticationBeValid, me)
	group.DELETE("/items/:id", handlers.MustBeAuthenticated, rights(false, []string{"manage_items"}), deleteItem)
})
```

The middlewares which authenticate need `jwt.secret`, see [API features](api-features.md).

**Web server.** The function is called once per locale, with the group of the locale (`/en-us`). `getRelativePath`
returns, for the key of a route of the routes file, its path relative to the group, whether the page is partial, the
middleware saving the key of the page in the context, and whether the route exists for the locale:

```go
server.Web.AddCustomRoutes(func(g *gin.RouterGroup, locale string, handlers route.Middlewares, getRelativePath func(key string) (string, bool, gin.HandlerFunc, bool)) {
	if path, _, savePageKey, ok := getRelativePath("index"); ok {
		g.GET(path, savePageKey, func(c *gin.Context) {
			page := html.Of(c).PageByKey(c, "index")
			page.Content.Title = page.Title
			page.RenderOk(c)
		})
	}
})
```
