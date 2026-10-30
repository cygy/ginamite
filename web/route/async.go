package route

import (
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/cygy/ginamite/common/log"
	apiResponse "github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/html"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// loadAsyncRoutes loads the passthrough async calls of the routes file for the locale into the group.
func (r *Router) loadAsyncRoutes(group *gin.RouterGroup, locale string) {
	for key, async := range html.Main.Async {
		path, ok := relativeAsyncPath(key, locale)
		if !ok {
			continue
		}

		method := strings.ToUpper(async.Method)
		switch method {
		case api.GET, api.POST, api.PUT, api.PATCH, api.DELETE:
			group.Handle(method, path, forwardToAPI(async.Target, method))
		default:
			log.WithFields(logrus.Fields{
				"async":  key,
				"method": async.Method,
			}).Warn("unsupported method for the async call")
		}
	}
}

// forwardToAPI returns the handler which sends the request to the endpoint target of the API with the method, and
// sends the response of the API back. The parameters of the route, prefixed by a colon in the target, are escaped.
func forwardToAPI(target, method string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		var res any

		request := api.NewRequest()
		request.Endpoint = expandEndpoint(target, c.Params, c.Request.URL.RawQuery)
		request.Method = method
		request.Body = body
		request.Response = &res
		request.Authenticated = true
		request.UseDefaultResponseHandlers = false
		request.ContentType = c.Request.Header.Get("Content-Type")

		result, _ := api.Main.Call(c, request)
		switch {
		case result.RequestError != nil:
			c.AbortWithStatus(http.StatusInternalServerError)
		case result.APIError.IsNil():
			c.AbortWithStatusJSON(result.StatusCode, res)
		default:
			c.AbortWithStatusJSON(result.StatusCode, struct {
				Error apiResponse.Error `json:"error"`
			}{
				Error: result.APIError,
			})
		}
	}
}

// expandEndpoint replaces the parameters of the target, prefixed by a colon, with their escaped value, and appends
// the query. The longest names are replaced first, so that ":id" does not alter ":identifier".
func expandEndpoint(target string, params gin.Params, rawQuery string) string {
	sorted := append(gin.Params{}, params...)
	sort.SliceStable(sorted, func(i, j int) bool { return len(sorted[i].Key) > len(sorted[j].Key) })

	replacements := make([]string, 0, 2*len(sorted))
	for _, param := range sorted {
		replacements = append(replacements, ":"+param.Key, url.PathEscape(param.Value))
	}

	endpoint := strings.NewReplacer(replacements...).Replace(target)
	if len(rawQuery) > 0 {
		endpoint += "?" + rawQuery
	}

	return endpoint
}
