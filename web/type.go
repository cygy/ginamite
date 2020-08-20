package web

import (
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/route"
)

// Server : definition of an web server.
type Server struct {
	RoutesFile string
	Functions  Handlers

	customRoutes []route.ConfigureCustomRoutes
}

// Handlers : list of the functions of an API.
type Handlers struct {
	RecurringTasks               func()
	APIMiddlewares               []api.ResponseHandlerFunc
	RouterConfigureDefaultRoutes route.ConfigureDefaultRoutes
}

// Pagination : properties of a Pagination struct.
type Pagination struct {
	CountOfPages int
	Pages        []int
	TotalItems   int
	ItemsPerPage int
	CurrentPage  int
	PreviousPage int
	NextPage     int
}
