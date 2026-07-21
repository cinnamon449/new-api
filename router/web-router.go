package router

import (
	"embed"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// ThemeAssets holds the embedded frontend assets for each shipped theme.
type ThemeAssets struct {
	DefaultBuildFS   embed.FS
	DefaultIndexPage []byte
	ClassicBuildFS   embed.FS
	ClassicIndexPage []byte
	InterapiBuildFS  embed.FS
	InterapiIndexPage []byte
}

func SetWebRouter(router *gin.Engine, assets ThemeAssets) {
	defaultFS := common.EmbedFolder(assets.DefaultBuildFS, "web/default/dist")
	classicFS := common.EmbedFolder(assets.ClassicBuildFS, "web/classic/dist")
	interapiFS := common.EmbedFolder(assets.InterapiBuildFS, "web/interapi/dist")
	themeFS := common.NewThemeAwareFS(defaultFS, map[string]static.ServeFileSystem{
		"classic":  classicFS,
		"interapi": interapiFS,
	})

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.GlobalWebRateLimit())
	router.Use(middleware.Cache())
	router.Use(static.Serve("/", themeFS))
	router.NoRoute(func(c *gin.Context) {
		c.Set(middleware.RouteTagKey, "web")
		if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") || strings.HasPrefix(c.Request.RequestURI, "/assets") {
			controller.RelayNotFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		var indexPage []byte
		switch common.GetTheme() {
		case "classic":
			indexPage = assets.ClassicIndexPage
		case "interapi":
			indexPage = assets.InterapiIndexPage
		default:
			indexPage = assets.DefaultIndexPage
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexPage)
	})
}
