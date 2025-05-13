package web

import (
	"NUMParser/config"
	"NUMParser/db"
	"NUMParser/db/models"
	"NUMParser/db/tmdb"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	route        *gin.Engine
	cachedMovies struct {
		sync.RWMutex
		movies     []*models.Entity
		lastUpdate time.Time
	}
	cacheDuration = 5 * time.Minute
)

func isAnimation(e *models.Entity) bool {
	for _, g := range e.Genres {
		if g == nil {
			continue
		}
		name := strings.ToLower(g.Name)
		if name == "мультфильм" || name == "animation" {
			return true
		}
	}
	return false
}

func compareQuality(qualityI, qualityJ string) bool {
	qualityMap := map[string]int{
		"4K":    4,
		"1080p": 3,
		"720p":  2,
		"SD":    1,
	}
	return qualityMap[qualityI] > qualityMap[qualityJ]
}

func updateMoviesCache() {
	entities := tmdb.GetAllMovies()
	var movies []*models.Entity
	for _, m := range entities {
		if m.MediaType == "movie" && !isAnimation(m) {
			categories := tmdb.GetReleaseCategoriesByTMDBID(m.ID)
			if categories == "Movie" {
				if m.GetTorrent() == nil {
					m.SetTorrent(&models.TorrentDetails{})
				}
				movies = append(movies, m)
			}
		}
	}

	// Предварительно получаем все необходимые данные
	for _, m := range movies {
		if m.GetTorrent() == nil {
			m.SetTorrent(&models.TorrentDetails{})
		}
		m.GetTorrent().CreateDate = tmdb.GetReleaseCreateDateByTMDBID(m.ID)
		m.GetTorrent().VideoQuality = getQualityValue(tmdb.GetReleaseQualityByTMDBID(m.ID))
	}

	sort.Slice(movies, func(i, j int) bool {
		if movies[i].GetTorrent().CreateDate.Equal(movies[j].GetTorrent().CreateDate) {
			return movies[i].GetTorrent().VideoQuality > movies[j].GetTorrent().VideoQuality
		}
		return movies[i].GetTorrent().CreateDate.After(movies[j].GetTorrent().CreateDate)
	})

	cachedMovies.Lock()
	cachedMovies.movies = movies
	cachedMovies.lastUpdate = time.Now()
	cachedMovies.Unlock()
}

func getQualityValue(quality string) int {
	qualityMap := map[string]int{
		"4K":    4,
		"1080p": 3,
		"720p":  2,
		"SD":    1,
	}
	return qualityMap[quality]
}

func getCachedMovies() []*models.Entity {
	cachedMovies.RLock()
	defer cachedMovies.RUnlock()

	if time.Since(cachedMovies.lastUpdate) > cacheDuration {
		go updateMoviesCache()
	}

	return cachedMovies.movies
}

func setupRouter() *gin.Engine {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.Static("/css", "public/css")
	r.Static("/img", "public/img")
	r.Static("/js", "public/js")
	r.StaticFile("/", "public/index.html")

	// Инициализация кэша при старте
	updateMoviesCache()

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

	// Фильмы (без мультфильмов)
	r.GET("/api/lampac/movies", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		movies := getCachedMovies()
		sendMoviesResponse(c, movies, page)
	})

	// Русские фильмы (без мультфильмов)
	r.GET("/api/lampac/russian_movies", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		entities := tmdb.GetAllMovies()
		var russianMovies []*models.Entity
		for _, m := range entities {
			if m.OriginalLanguage == "ru" && m.MediaType == "movie" && !isAnimation(m) {
				russianMovies = append(russianMovies, m)
			}
		}
		sort.Slice(russianMovies, func(i, j int) bool {
			return russianMovies[i].UpdateDate.After(russianMovies[j].UpdateDate)
		})
		sendMoviesResponse(c, russianMovies, page)
	})

	// Мультфильмы (только фильмы)
	r.GET("/api/lampac/cartoons", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		entities := tmdb.GetAllMovies()
		var cartoons []*models.Entity
		for _, m := range entities {
			if m.MediaType == "movie" && isAnimation(m) {
				cartoons = append(cartoons, m)
			}
		}
		sort.Slice(cartoons, func(i, j int) bool {
			return cartoons[i].UpdateDate.After(cartoons[j].UpdateDate)
		})
		sendMoviesResponse(c, cartoons, page)
	})

	// Мультсериалы (только сериалы)
	r.GET("/api/lampac/cartoons_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		entities := tmdb.GetAllTV()
		var cartoonsTV []*models.Entity
		for _, m := range entities {
			if isAnimation(m) {
				cartoonsTV = append(cartoonsTV, m)
			}
		}
		sort.Slice(cartoonsTV, func(i, j int) bool {
			return cartoonsTV[i].UpdateDate.After(cartoonsTV[j].UpdateDate)
		})
		sendMoviesResponse(c, cartoonsTV, page)
	})

	// Сериалы (без мультсериалов)
	r.GET("/api/lampac/all_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		entities := tmdb.GetAllTV()
		var allTV []*models.Entity
		for _, m := range entities {
			if !isAnimation(m) {
				allTV = append(allTV, m)
			}
		}
		sort.Slice(allTV, func(i, j int) bool {
			return allTV[i].UpdateDate.After(allTV[j].UpdateDate)
		})
		sendMoviesResponse(c, allTV, page)
	})

	// Русские сериалы (без мультсериалов)
	r.GET("/api/lampac/russian_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		entities := tmdb.GetAllTV()
		var russianTV []*models.Entity
		for _, m := range entities {
			if m.OriginalLanguage == "ru" && !isAnimation(m) {
				russianTV = append(russianTV, m)
			}
		}
		sort.Slice(russianTV, func(i, j int) bool {
			return russianTV[i].UpdateDate.After(russianTV[j].UpdateDate)
		})
		sendMoviesResponse(c, russianTV, page)
	})

	return r
}

// Вспомогательная функция для получения параметра страницы
func getPageParam(c *gin.Context) int {
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	return page
}

// Вспомогательная функция для отправки ответа с фильмами
func sendMoviesResponse(c *gin.Context, movies []*models.Entity, page int) {
	totalResults := len(movies)
	totalPages := (totalResults + 19) / 20 // Округляем вверх

	// Вычисляем индексы для текущей страницы
	start := (page - 1) * 20
	end := start + 20
	if end > totalResults {
		end = totalResults
	}

	// Получаем только фильмы для текущей страницы
	var results []map[string]interface{}
	for _, m := range movies[start:end] {
		releaseQuality := tmdb.GetReleaseQualityByTMDBID(m.ID)
		categories := tmdb.GetReleaseCategoriesByTMDBID(m.ID)
		createDate := tmdb.GetReleaseCreateDateByTMDBID(m.ID)
		poster := m.PosterPath
		if poster != "" {
			poster = strings.TrimPrefix(poster, "http://image.tmdb.org")
			poster = strings.TrimPrefix(poster, "https://image.tmdb.org")
			poster = "https://torrs.ru/tmdbimg" + poster
		}

		backdrop := m.BackdropPath
		if backdrop != "" {
			backdrop = strings.TrimPrefix(backdrop, "http://image.tmdb.org")
			backdrop = strings.TrimPrefix(backdrop, "https://image.tmdb.org")
			backdrop = "https://torrs.ru/tmdbimg" + backdrop
		}

		name := m.Name
		if name == "" {
			name = m.Title
		}
		if name == "" {
			name = m.OriginalName
		}
		if name == "" {
			name = m.OriginalTitle
		}

		original_name := m.OriginalName
		if original_name == "" {
			original_name = m.OriginalTitle
		}
		if original_name == "" {
			original_name = name
		}

		first_air_date := m.FirstAirDate
		if t, err := time.Parse("02.01.2006", m.FirstAirDate); err == nil {
			first_air_date = t.Format("2006-01-02")
		} else {
			first_air_date = ""
		}

		releaseDate := m.ReleaseDate
		if t, err := time.Parse("02.01.2006", m.ReleaseDate); err == nil {
			releaseDate = t.Format("2006-01-02")
		}

		results = append(results, map[string]interface{}{
			"backdrop_path":     m.BackdropPath,
			"first_air_date":    first_air_date,
			"id":                m.ID,
			"name":              name,
			"number_of_seasons": m.NumberOfSeasons,
			"seasons":           m.Seasons,
			"original_name":     original_name,
			"overview":          m.Overview,
			"poster_path":       m.PosterPath,
			"release_date":      releaseDate,
			"still_path":        "",
			"vote_average":      m.VoteAverage,
			"vote_count":        m.VoteCount,
			"source":            "Lampa",
			"original_language": m.OriginalLanguage,
			"media_type":        m.MediaType,
			"video":             m.Video,
			"update_date":       m.UpdateDate,
			"release_quality":   releaseQuality,
			"categories":        categories,
			"create_date":       createDate,
		})
	}

	c.JSON(200, gin.H{
		"page":          page,
		"results":       results,
		"total_pages":   totalPages,
		"total_results": totalResults,
	})
}

var isSetStatic bool

func SetStaticReleases() {
	if !isSetStatic {
		route.Static("/releases", config.SaveReleasePath)
		isSetStatic = true
	}
}

func Start(port string) {
	route = gin.Default()
	go func() {
		route = setupRouter()
		err := route.Run(":" + port)
		if err != nil {
			log.Println("Error start web server on port", port, ":", err)
		}
	}()
}
