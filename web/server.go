package web

import (
	"NUMParser/config"
	"NUMParser/db"
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed public/css public/js public/img public/index.html public/settings.html
var publicFS embed.FS

func subFS(prefix string) http.FileSystem {
	sub, err := fs.Sub(publicFS, prefix)
	if err != nil {
		log.Fatalf("embed sub %q: %v", prefix, err)
	}
	return http.FS(sub)
}

func mustReadEmbed(name string) []byte {
	buf, err := publicFS.ReadFile(name)
	if err != nil {
		log.Fatalf("embed read %q: %v", name, err)
	}
	return buf
}

var route *gin.Engine
var currentPort string

func setupRouter() *gin.Engine {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.StaticFS("/css", subFS("public/css"))
	r.StaticFS("/img", subFS("public/img"))
	r.StaticFS("/js", subFS("public/js"))
	r.Static("/releases", config.SaveReleasePath)

	indexHTML := mustReadEmbed("public/index.html")
	settingsHTML := mustReadEmbed("public/settings.html")
	htmlType := "text/html; charset=utf-8"
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, htmlType, indexHTML)
	})
	r.GET("/settings", func(c *gin.Context) {
		c.Data(http.StatusOK, htmlType, settingsHTML)
	})
	r.GET("/settings/", func(c *gin.Context) {
		c.Data(http.StatusOK, htmlType, settingsHTML)
	})

	// http://127.0.0.1:38888/search?query=venom
	r.GET("/search", func(c *gin.Context) {
		if query, ok := c.GetQuery("query"); ok {
			torrs := db.SearchTorr(query)
			c.JSON(200, torrs)
			return
		}
		c.Status(http.StatusBadRequest)
		return
	})

	r.GET("/api/settings", func(c *gin.Context) {
		settings, err := config.LoadConfig()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"settings":         settings,
			"restart_required": currentPort != "" && settings.Port != currentPort,
		})
	})

	r.POST("/api/settings", func(c *gin.Context) {
		var settings config.ConfigParser
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settings payload"})
			return
		}

		saved, err := config.SaveConfig(settings)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"settings":         saved,
			"restart_required": currentPort != "" && saved.Port != currentPort,
		})
	})

	return r
}

func Start(port string) {
	log.Println("Init web")
	go func() {
		currentPort = port
		route = setupRouter()
		addr := "0.0.0.0:" + port
		log.Println("Start web server on", addr)
		err := route.Run(addr)
		if err != nil {
			log.Println("Error start web server on port", port, ":", err)
		}
	}()
}
