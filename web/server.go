package web

import (
	"NUMParser/config"
	"NUMParser/db"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var route *gin.Engine
var currentPort string

func setupRouter() *gin.Engine {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.Static("/css", "public/css")
	r.Static("/img", "public/img")
	r.Static("/js", "public/js")
	r.StaticFile("/", "public/index.html")
	r.StaticFile("/settings", "public/settings.html")
	r.StaticFile("/settings/", "public/settings.html")

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

var isSetStatic bool

func SetStaticReleases() {
	if !isSetStatic {
		route.Static("/releases", config.SaveReleasePath)
		isSetStatic = true
	}
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
