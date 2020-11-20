package html

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/common/template/html"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	// LocaleKeyword key of the variable locale.
	LocaleKeyword = "Locale"
	// WebHostKeyword key of the variable webhost.
	WebHostKeyword = "WebHost"
	// StaticHostKeyword key of the variable statichost.
	StaticHostKeyword = "StaticHost"
	// APIHostKeyword key of the variable apihost.
	APIHostKeyword = "ApiHost"

	// The timezone locations.
	locations map[string]*time.Location
)

// Page struct of a HTML page.
type Page struct {
	Locale        string `yaml:"-"` // Defined by the constructor
	Timezone      string `yaml:"-"` // Defined by the constructor
	Paths         map[string]string
	Template      PageTemplate
	AccessRight   AccessRight `yaml:"access"`
	ScriptFiles   []Script    `yaml:"scripts"`
	Meta          Meta
	Title         string
	Sitemap       Sitemap
	Configuration PageConfiguration `yaml:"-"`       // Defined by the constructor
	Content       PageContent       `yaml:"-"`       // Defined by the each script
	Partial       bool              `yaml:"partial"` // true if this page is a partial page (loaded by an ajax call), false if it is a full page.
	Key           string            // The key defining the route in the 'routes' file.
	BetaVersion   bool
	Production    bool
	Vars          map[string]interface{} // Common variables to all the pages.
}

// NewPage : returns a new struct *Page
func NewPage() *Page {
	page := &Page{}
	page.Content.Vars = map[string]interface{}{}
	page.Vars = map[string]interface{}{}
	page.Partial = false

	return page
}

// Copy : copies a struct into an other.
func (page *Page) Copy() (copied *Page) {
	copied = &Page{}
	copied.Locale = page.Locale
	copied.Timezone = page.Timezone
	copied.Template = page.Template
	copied.AccessRight = page.AccessRight
	copied.Sitemap = page.Sitemap
	copied.Meta = *page.Meta.Copy()
	copied.Title = page.Title
	copied.Configuration = page.Configuration
	copied.Partial = page.Partial
	copied.Key = page.Key
	copied.BetaVersion = page.BetaVersion
	copied.Production = page.Production

	copied.Paths = make(map[string]string, len(page.Paths))
	for key, value := range page.Paths {
		copied.Paths[key] = value
	}

	copied.Vars = make(map[string]interface{}, len(page.Vars))
	for key, value := range page.Vars {
		copied.Vars[key] = value
	}

	copied.ScriptFiles = make([]Script, len(page.ScriptFiles))
	copy(copied.ScriptFiles, page.ScriptFiles)

	return
}

// Merge : merges two Page structs.
func (page *Page) Merge(toMerge Page) {
	if len(toMerge.Locale) > 0 {
		page.Locale = toMerge.Locale
	}
	if len(toMerge.Timezone) > 0 {
		page.Timezone = toMerge.Timezone
	}
	if len(toMerge.AccessRight) > 0 {
		page.AccessRight = toMerge.AccessRight
	}
	if len(toMerge.Title) > 0 {
		page.Title = toMerge.Title
	}

	page.Template.Merge(toMerge.Template)
	page.Meta.Merge(toMerge.Meta)

	if toMerge.Sitemap.IsDefined() {
		page.Sitemap = toMerge.Sitemap
	}

	if len(toMerge.Paths) > 0 {
		if page.Paths == nil {
			page.Paths = make(map[string]string, len(toMerge.Paths))
		}
		for key, value := range toMerge.Paths {
			page.Paths[key] = value
		}
	}

	if len(toMerge.Vars) > 0 {
		if page.Vars == nil {
			page.Vars = make(map[string]interface{}, len(toMerge.Vars))
		}
		for key, value := range toMerge.Vars {
			page.Vars[key] = value
		}
	}

	if toMerge.ScriptFiles != nil {
		if page.ScriptFiles == nil {
			page.ScriptFiles = []Script{}
		}
		page.ScriptFiles = append(page.ScriptFiles, toMerge.ScriptFiles...)
	}

	page.Partial = toMerge.Partial
	page.Key = toMerge.Key
}

// UpdateMetaRobots : updates the meta 'robots' accordingl to the sitemap's properties.
func (page *Page) UpdateMetaRobots() {
	robots := strings.ToLower(page.Meta.Robots)

	if page.Sitemap.Included {
		if strings.Contains(robots, "noindex") {
			robots = strings.Replace(robots, "noindex", "index", -1)
		} else if len(robots) == 0 {
			robots = "index"
		} else if !strings.Contains(robots, "index") {
			robots = fmt.Sprintf("%s,index", robots)
		}
	} else if !strings.Contains(robots, "noindex") {
		if strings.Contains(robots, "index") {
			robots = strings.Replace(robots, "index", "noindex", -1)
		} else if len(robots) == 0 {
			robots = "noindex"
		} else {
			robots = fmt.Sprintf("%s,noindex", robots)
		}
	}

	page.Meta.Robots = robots
}

// PreparePaths : sets up the global page struct.
func (page *Page) PreparePaths(webHost string, applicationLocales []string, fallbackLocales map[string]string) {
	// Remove the unsupported locales from the paths map.
	paths := map[string]string{}
	localeKey := fmt.Sprintf("{{.%s}}", LocaleKeyword)
	for _, applicationLocale := range applicationLocales {
		if path, ok := page.Paths[applicationLocale]; ok {
			paths[applicationLocale] = strings.Replace(path, localeKey, applicationLocale, 1)
		}
	}
	page.Paths = paths

	// Sets up the href lang to the metadata.
	page.Meta.setUpHrefLangs(page.Paths, fallbackLocales, webHost)
}

// PrepareWithLocale : sets up a specific page struct before added some keywords.
func (page *Page) PrepareWithLocale(currentLocale, currentTimezone, webHost, staticHost, apiHost string, supportedLocales []string) {
	// Save the locale and the timezone.
	page.Locale = currentLocale
	page.Timezone = currentTimezone

	// Localize some strings.
	page.Title = page.rawString(page.Title)

	page.Meta.Title = page.rawString(page.Meta.Title)
	page.Meta.Description = page.rawString(page.Meta.Description)
	page.Meta.OpenGraph.Title = page.rawString(page.Meta.OpenGraph.Title)
	page.Meta.OpenGraph.Description = page.rawString(page.Meta.OpenGraph.Description)
	page.Meta.OpenGraph.SiteName = page.rawString(page.Meta.OpenGraph.SiteName)
	page.Meta.Twitter.ImageAlt = page.rawString(page.Meta.Twitter.ImageAlt)

	// Localize the meta keywords.
	for i, keyword := range page.Meta.Keywords {
		page.Meta.Keywords[i] = page.rawString(keyword)
	}

	// Replace the variables to the URLs.
	variables := map[string]string{
		WebHostKeyword:    webHost,
		StaticHostKeyword: staticHost,
		APIHostKeyword:    apiHost,
		LocaleKeyword:     currentLocale,
	}
	for name, value := range variables {
		key := fmt.Sprintf("{{.%s}}", name)
		page.Meta.ImageSource = strings.Replace(page.Meta.ImageSource, key, value, 1)
		page.Meta.OpenGraph.Image = strings.Replace(page.Meta.OpenGraph.Image, key, value, 1)
		page.Meta.OpenGraph.URL = strings.Replace(page.Meta.OpenGraph.URL, key, value, 1)

		for i, icon := range page.Meta.Icons {
			icon.Href = strings.Replace(icon.Href, key, value, 1)
			page.Meta.Icons[i] = icon
		}
		for i, alternate := range page.Meta.Alternates {
			alternate.Href = strings.Replace(alternate.Href, key, value, 1)
			page.Meta.Alternates[i] = alternate
		}
		for i, stylesheet := range page.Meta.Stylesheets {
			stylesheet.Href = strings.Replace(stylesheet.Href, key, value, 1)
			page.Meta.Stylesheets[i] = stylesheet
		}
		for i, scriptFile := range page.ScriptFiles {
			scriptFile.Src = strings.Replace(scriptFile.Src, key, value, 1)
			page.ScriptFiles[i] = scriptFile
		}
	}

	// Set the locales up in the open graph tags.
	page.Meta.OpenGraph.Locale = FormatOpenGraphLocale(currentLocale)
	page.Meta.OpenGraph.AlternateLocales = []string{}
	for _, locale := range supportedLocales {
		if locale != currentLocale {
			page.Meta.OpenGraph.AlternateLocales = append(page.Meta.OpenGraph.AlternateLocales, FormatOpenGraphLocale(locale))
		}
	}
}

// InitializeData : initializes the data into the page struct before added some keywords.
func (page *Page) InitializeData(c *gin.Context) {
	// Sets up the userID and the username.
	page.Content = PageContent{}
	page.Content.Title = page.Title

	for key, value := range page.Vars {
		page.SetVar(key, value)
	}

	if tokenString, err := c.Cookie(jwtCookieName); err == nil {
		if token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(jwtSecret), nil
		}); err == nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				page.Content.CurrentUser.ID = claims["sub"].(string)
				page.Content.CurrentUser.Name = claims["nickname"].(string)
				page.Content.CurrentUser.Image1x = claims["img1"].(string)
				page.Content.CurrentUser.Image2x = claims["img2"].(string)
				if unread, ok := claims["unread"]; ok {
					page.Content.CurrentUser.Unread = uint(unread.(float64))
				} else {
					page.Content.CurrentUser.Unread = 0
				}

				timezone := claims["timezone"].(string)
				if len(timezone) > 0 {
					page.Timezone = timezone
				}
			}
		}
	}
}

// Process : replaces the variables into the target properties before processing the templates.
func (page *Page) Process() {
	page.Content.Title = page.replaceVars(page.Content.Title)

	page.Meta.Title = page.replaceVars(page.Meta.Title)
	page.Meta.Description = page.replaceVars(page.Meta.Description)
	page.Meta.Twitter.ImageAlt = page.replaceVars(page.Meta.Twitter.ImageAlt)

	for i, keyword := range page.Meta.Keywords {
		page.Meta.Keywords[i] = page.replaceVars(keyword)
	}

	for i, hrefLang := range page.Meta.HrefLang {
		page.Meta.HrefLang[i].Href = page.replaceVars(hrefLang.Href)
	}

	// Choose the right canonical URL.
	page.Meta.setUpCanonicalURL(page.Locale)

	// Set the OpenGraph tags up.
	page.Meta.OpenGraph.URL = page.Meta.CanonicalURL
	page.Meta.OpenGraph.SiteName = page.replaceVars(page.Meta.OpenGraph.SiteName)

	if len(page.Meta.OpenGraph.Title) == 0 {
		page.Meta.OpenGraph.Title = page.Meta.Title
	} else {
		page.Meta.OpenGraph.Title = page.replaceVars(page.Meta.OpenGraph.Title)
	}

	if len(page.Meta.OpenGraph.Description) == 0 {
		page.Meta.OpenGraph.Description = page.Meta.Description
	} else {
		page.Meta.OpenGraph.Description = page.replaceVars(page.Meta.OpenGraph.Description)
	}

	// Parse the main template: it will be included to the decorator template later.
	var output bytes.Buffer
	if err := html.Main.ExecuteTemplate(&output, page.Template.Main, page); err != nil {
		log.WithFields(logrus.Fields{
			"template": page.Template.Main,
			"vars":     page.Content,
			"error":    err.Error(),
		}).Error("unable to execute the template")
	} else {
		page.Content.parsedMainTemplate = output.String()
	}
}

// RenderOk : renders a HTML page with a HTTP code 200.
func (page *Page) RenderOk(c *gin.Context) {
	page.render(c, http.StatusOK)
}

// RenderError : renders a HTML page with a HTTP code 500.
func (page *Page) RenderError(c *gin.Context) {
	page.render(c, http.StatusInternalServerError)
}

// RenderUnauthorized : renders a HTML page with a HTTP code 401.
func (page *Page) RenderUnauthorized(c *gin.Context) {
	page.render(c, http.StatusUnauthorized)
}

// RenderNotFound : renders a HTML page with a HTTP code 404.
func (page *Page) RenderNotFound(c *gin.Context) {
	page.render(c, http.StatusNotFound)
}

// SetError : sets up the error of the page.
func (page *Page) SetError(err response.Error) {
	page.Content.Error.Message = err.Message
	page.Content.Error.Reason = err.Reason
	page.Content.Error.Recovery = err.Recovery
}

// SetInfo : set up the error of the page.
func (page *Page) SetInfo(info Info) {
	page.Content.Info.Title = info.Title
	page.Content.Info.Message = info.Message
}
