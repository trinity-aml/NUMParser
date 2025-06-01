package web

import (
	"NUMParser/config"
	"NUMParser/db/models"
	"NUMParser/db/tmdb"
	"NUMParser/utils"
	"compress/gzip"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	route        *gin.Engine
	cachedMovies struct {
		sync.RWMutex
		movies4kNew      []*models.Entity
		movies4k         []*models.Entity
		moviesNew        []*models.Entity
		movies           []*models.Entity
		moviesRuNew      []*models.Entity
		moviesRu         []*models.Entity
		tvShowNew        []*models.Entity
		tvShow           []*models.Entity
		allTVShows       []*models.Entity
		tvShowRuNew      []*models.Entity
		tvShowRu         []*models.Entity
		allTVShowsRu     []*models.Entity
		cartoonMovies    []*models.Entity
		cartoonMoviesNew []*models.Entity
		allCartoonMovies []*models.Entity
		cartoonSeries    []*models.Entity
		cartoonSeriesNew []*models.Entity
		allCartoonSeries []*models.Entity
		lastUpdate       time.Time
	}
	// cacheDuration = 5 * time.Minute
)

// SaveLampacData сохраняет данные в файл с префиксом lampac_
func SaveLampacData(baseName string, data interface{}) {
	// Формируем полное имя файла
	filename := "lampac_" + baseName + ".json"
	fullPath := filepath.Join(config.SaveReleasePath, filename)

	// Создаем директорию
	if err := os.MkdirAll(config.SaveReleasePath, 0777); err != nil {
		log.Printf("Ошибка создания директории: %v", err)
		return
	}

	// Создаем файл
	file, err := os.Create(fullPath)
	if err != nil {
		log.Printf("Ошибка создания файла: %v", err)
		return
	}
	defer file.Close()

	// Сжимаем в GZIP
	gz := gzip.NewWriter(file)
	defer gz.Close()

	// Кодируем в JSON
	if err := json.NewEncoder(gz).Encode(data); err != nil {
		log.Printf("Ошибка кодирования JSON: %v", err)
		return
	}

	log.Printf("Сохранено: %s", fullPath)
}

func UpdateMoviesCache() {

	// Получаем все данные
	movieEntities := tmdb.GetAllMovies()
	tvEntities := tmdb.GetAllTV()

	// // Фильтруем данные по категориям
	moviesRuNew, moviesRu := filterEntitiesByCategory(movieEntities, []string{models.CatMovie}, "ru", 2, 200, "hd")
	moviesNew, movies := filterEntitiesByCategory(movieEntities, []string{models.CatMovie}, "notru", 2, 200, "hd")
	tvShowRuNew, tvShowRu := filterEntitiesByCategory(tvEntities, []string{models.CatSeries}, "ru", 2, 200, "all")
	tvShowNew, tvShow := filterEntitiesByCategory(tvEntities, []string{models.CatSeries}, "notru", 2, 200, "all")
	cartoonMoviesNew, cartoonMovies := filterEntitiesByCategory(movieEntities, []string{models.CatCartoonMovie}, "all", 2, 200, "all")
	cartoonSeriesNew, cartoonSeries := filterEntitiesByCategory(tvEntities, []string{models.CatCartoonSeries}, "all", 2, 200, "all")
	//animeNew, anime := filterEntitiesByCategory(movieEntities, []string{models.CatAnime}, "all", 2, 200, "all")
	movies4kNew, movies4k := filterEntitiesByCategory(movieEntities, []string{models.CatMovie}, "all", 4, 300, "4k")
	sortEntities := func(entities []*models.Entity) {
		sort.Slice(entities, func(i, j int) bool {
			ti := entities[i].GetTorrent()
			tj := entities[j].GetTorrent()
			if ti == nil && tj == nil {
				return false
			}
			if ti == nil {
				return false
			}
			if tj == nil {
				return true
			}
			if ti.CreateDate.Equal(tj.CreateDate) {
				return ti.VideoQuality > tj.VideoQuality
			}
			return ti.CreateDate.After(tj.CreateDate)
		})
	}

	// Функция сортировки по дате релиза
	sortReleaseDate := func(entities []*models.Entity) {
		sort.Slice(entities, func(i, j int) bool {
			// Парсим даты
			dateI, errI := time.Parse("02.01.2006", entities[i].ReleaseDate)
			dateJ, errJ := time.Parse("02.01.2006", entities[j].ReleaseDate)

			// Если не удалось распарсить — ставим минимальное значение
			if errI != nil {
				dateI = time.Time{}
			}
			if errJ != nil {
				dateJ = time.Time{}
			}

			// Сортировка по дате релиза
			if dateI.Equal(dateJ) {
				torrI := entities[i].GetTorrent()
				torrJ := entities[j].GetTorrent()
				// Проверяем на nil
				if torrI == nil && torrJ == nil {
					return false
				}
				if torrI == nil {
					return false
				}
				if torrJ == nil {
					return true
				}
				return torrI.VideoQuality > torrJ.VideoQuality
			}
			return dateI.After(dateJ)
		})
	}

	// Сортируем все списки по дате добавления торрента
	sortEntities(movies4kNew)
	sortEntities(moviesNew)
	sortEntities(moviesRuNew)
	sortEntities(tvShowNew)
	sortEntities(tvShowRuNew)
	sortEntities(cartoonMoviesNew)
	sortEntities(cartoonSeriesNew)

	// Сортируем все списки по дате релиза
	sortReleaseDate(movies4k)
	sortReleaseDate(movies)
	sortReleaseDate(moviesRu)
	sortReleaseDate(tvShowRu)
	sortReleaseDate(tvShow)
	sortReleaseDate(cartoonMovies)
	sortReleaseDate(cartoonSeries)

	// Объединение списков
	allTVShows := make([]*models.Entity, 0, len(tvShowNew)+len(tvShow))
	allTVShows = append(allTVShows, tvShowNew...)
	allTVShows = append(allTVShows, tvShow...)

	allTVShowsRu := make([]*models.Entity, 0, len(tvShowRuNew)+len(tvShowRu))
	allTVShowsRu = append(allTVShowsRu, tvShowRuNew...)
	allTVShowsRu = append(allTVShowsRu, tvShowRu...)

	allCartoonMovies := make([]*models.Entity, 0, len(cartoonMoviesNew)+len(cartoonMovies))
	allCartoonMovies = append(allCartoonMovies, cartoonMoviesNew...)
	allCartoonMovies = append(allCartoonMovies, cartoonMovies...)

	allCartoonSeries := make([]*models.Entity, 0, len(cartoonSeriesNew)+len(cartoonSeries))
	allCartoonSeries = append(allCartoonSeries, cartoonSeriesNew...)
	allCartoonSeries = append(allCartoonSeries, cartoonSeries...)

	utils.FreeOSMemGC()

	// Сохраняем ВСЕ категории
	SaveLampacData("movies_ru_new", moviesRuNew)
	SaveLampacData("movies_ru", moviesRu)
	SaveLampacData("movies_new", moviesNew)
	SaveLampacData("movies", movies)
	SaveLampacData("tv_ru_new", tvShowRuNew)
	SaveLampacData("tv_ru", tvShowRu)
	SaveLampacData("tv_new", tvShowNew)
	SaveLampacData("tv", tvShow)
	SaveLampacData("cartoon_movies_new", cartoonMoviesNew)
	SaveLampacData("cartoon_movies", cartoonMovies)
	SaveLampacData("cartoon_series_new", cartoonSeriesNew)
	SaveLampacData("cartoon_series", cartoonSeries)
	//SaveLampacData("anime_new", animeNew)
	//SaveLampacData("anime", anime)
	SaveLampacData("movies_4k_new", movies4kNew)
	SaveLampacData("movies_4k", movies4k)
	SaveLampacData("all_tv_shows", allTVShows)
	SaveLampacData("all_tv_shows_ru", allTVShowsRu)
	SaveLampacData("all_cartoon_movies", allCartoonMovies)
	SaveLampacData("all_cartoon_series", allCartoonSeries)

	// Блокируем и обновляем кэш
	cachedMovies.Lock()
	defer cachedMovies.Unlock()

	// Обновляем кэш
	cachedMovies.movies4kNew = movies4kNew
	cachedMovies.movies4k = movies4k
	cachedMovies.moviesNew = moviesNew
	cachedMovies.movies = movies
	cachedMovies.moviesRuNew = moviesRuNew
	cachedMovies.moviesRu = moviesRu
	cachedMovies.tvShowRu = tvShowRu
	cachedMovies.tvShowRuNew = tvShowRuNew
	cachedMovies.allTVShowsRu = allTVShowsRu
	cachedMovies.tvShow = tvShow
	cachedMovies.tvShowNew = tvShowNew
	cachedMovies.allTVShows = allTVShows
	cachedMovies.cartoonMoviesNew = cartoonMoviesNew
	cachedMovies.cartoonMovies = cartoonMovies
	cachedMovies.allCartoonMovies = allCartoonMovies
	cachedMovies.cartoonSeriesNew = cartoonSeriesNew
	cachedMovies.cartoonSeries = cartoonSeries
	cachedMovies.allCartoonSeries = allCartoonSeries
	cachedMovies.lastUpdate = time.Now()
}

// Добавляем тип для возврата всех категорий
type CachedMoviesResponse struct {
	Movies4k         []*models.Entity
	Movies4kNew      []*models.Entity
	MoviesNew        []*models.Entity
	Movies           []*models.Entity
	MoviesRuNew      []*models.Entity
	MoviesRu         []*models.Entity
	TVShowNew        []*models.Entity
	TVShow           []*models.Entity
	AllTVShows       []*models.Entity
	TVShowRuNew      []*models.Entity
	TVShowRu         []*models.Entity
	AllTVShowsRu     []*models.Entity
	CartoonMovies    []*models.Entity
	CartoonMoviesNew []*models.Entity
	AllCartoonMovies []*models.Entity
	CartoonSeries    []*models.Entity
	CartoonSeriesNew []*models.Entity
	AllCartoonSeries []*models.Entity
	LastUpdate       time.Time
}

// Получаем все категории фильмов
func GetCachedMovies() CachedMoviesResponse {
	cachedMovies.RLock()
	defer cachedMovies.RUnlock()

	return CachedMoviesResponse{
		Movies4k:         cachedMovies.movies4k,
		Movies4kNew:      cachedMovies.movies4kNew,
		MoviesNew:        cachedMovies.moviesNew,
		Movies:           cachedMovies.movies,
		MoviesRuNew:      cachedMovies.moviesRuNew,
		MoviesRu:         cachedMovies.moviesRu,
		TVShowNew:        cachedMovies.tvShowNew,
		TVShow:           cachedMovies.tvShow,
		AllTVShows:       cachedMovies.allTVShows,
		TVShowRuNew:      cachedMovies.tvShowRuNew,
		TVShowRu:         cachedMovies.tvShowRu,
		AllTVShowsRu:     cachedMovies.allTVShowsRu,
		CartoonMovies:    cachedMovies.cartoonMovies,
		CartoonMoviesNew: cachedMovies.cartoonMoviesNew,
		AllCartoonMovies: cachedMovies.allCartoonMovies,
		CartoonSeries:    cachedMovies.cartoonSeries,
		CartoonSeriesNew: cachedMovies.cartoonSeriesNew,
		AllCartoonSeries: cachedMovies.allCartoonSeries,
		LastUpdate:       cachedMovies.lastUpdate,
	}
}

func filterEntitiesByCategory(
	entities []*models.Entity,
	categories []string,
	langMode string, // "split", "all", "ru", "notru"
	yearDelta int,
	minQuality int,
	qualityMode string, // "4k", "hd", "all"
) (newList, allList []*models.Entity) {
	categorySet := make(map[string]struct{})
	for _, cat := range categories {
		categorySet[cat] = struct{}{}
	}

	for _, m := range entities {
		torr := tmdb.GetTorrentDetailsByTMDBID(m.ID)
		if torr == nil {
			continue
		}
		m.SetTorrent(torr)

		if _, ok := categorySet[torr.Categories]; !ok {
			continue
		}

		isRu := m.OriginalLanguage == "ru"
		use := false
		switch langMode {
		case "split":
			use = true
		case "all":
			use = true
		case "ru":
			use = isRu
		case "notru":
			use = !isRu
		}

		if !use {
			continue
		}

		// Проверяем качество в зависимости от режима
		qualityOK := false
		switch qualityMode {
		case "4k":
			qualityOK = torr.VideoQuality >= 300 // 4K
		case "hd":
			qualityOK = torr.VideoQuality >= minQuality && torr.VideoQuality < 300 // HD
		case "all":
			qualityOK = torr.VideoQuality >= minQuality // Все что выше minQuality
		}

		if !qualityOK {
			continue
		}

		// Проверяем год
		yearOK := utils.Abs(torr.Year-time.Now().Year()) < yearDelta

		if yearOK {
			newList = append(newList, m)
		} else {
			allList = append(allList, m)
		}
	}

	return
}

func InitLampacRoutes(r *gin.RouterGroup) {

	UpdateMoviesCache()
	// utils.FreeOSMemGC() // Освобождаем память после инициализаци

	// Фильмы в высоком качестве новинки
	r.GET("/4k_new", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.Movies4kNew, page)
	})

	// Фильмы в высоком качестве
	r.GET("/4k", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.Movies4k, page)
	})

	// Новые фильмы
	r.GET("/movies_new", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.MoviesNew, page)
	})

	// Новые русские фильмы
	r.GET("/movies_ru_new", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.MoviesRuNew, page)
	})

	// Фильмы
	r.GET("/movies", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.Movies, page)
	})

	// Русские фильмы (без мультфильмов)
	r.GET("/movies_ru", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.MoviesRu, page)
	})

	// Мультфильмы (только фильмы)
	r.GET("/cartoons", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.AllCartoonMovies, page)
	})

	// Мультсериалы (только сериалы)
	r.GET("/cartoons_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.AllCartoonSeries, page)
	})

	// Сериалы (без мультсериалов)
	r.GET("/all_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.AllTVShows, page)
	})

	// Русские сериалы (без мультсериалов)
	r.GET("/russian_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.AllTVShowsRu, page)
	})

}

// Вспомогательная функция отображения качества
func getQualityText(videoQuality int) string {
	switch {
	// SD (0-99)
	case videoQuality >= 0 && videoQuality <= 99:
		return "SD"

	// 720p (100-199)
	case videoQuality == 100:
		return "WEBDL 720p"
	case videoQuality == 101:
		return "BDRip 720p"
	case videoQuality >= 102 && videoQuality <= 199:
		return "BDRip HEVC 720p"

	// 1080p (200-299)
	case videoQuality == 200:
		return "WEBDL 1080p"
	case videoQuality == 201:
		return "BDRip 1080p"
	case videoQuality == 202:
		return "BDRip HEVC 1080p"
	case videoQuality == 203:
		return "Remux 1080p"
	case videoQuality >= 204 && videoQuality <= 299:
		return "1080p" // другие 1080p

	// 2160p (300+)
	case videoQuality == 300:
		return "WEBDL 2160p"
	case videoQuality == 301:
		return "WEBDL HDR 2160p"
	case videoQuality == 302:
		return "WEBDL DV 2160p"
	case videoQuality == 303:
		return "BDRip 2160p"
	case videoQuality == 304:
		return "BDRip HDR 2160p"
	case videoQuality == 305:
		return "BDRip DV 2160p"
	case videoQuality == 306:
		return "Remux 2160p"
	case videoQuality == 307:
		return "Remux HDR 2160p"
	case videoQuality == 308:
		return "Remux DV 2160p"
	case videoQuality >= 309:
		return "2160p" // другие 4K

	default:
		return ""
	}
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
		torr := tmdb.GetTorrentDetailsByTMDBID(m.ID)

		// Преобразуем VideoQuality в текстовый формат
		qualityText := getQualityText(torr.VideoQuality)

		releaseDate := m.ReleaseDate
		if t, err := time.Parse("02.01.2006", m.ReleaseDate); err == nil {
			releaseDate = t.Format("2006-01-02")
		}

		results = append(results, map[string]interface{}{
			"backdrop_path":     m.BackdropPath,
			"first_air_date":    m.FirstAirDate,
			"last_air_date":     m.LastAirDate,
			"id":                m.ID,
			"name":              m.Title,
			"number_of_seasons": m.NumberOfSeasons,
			"seasons":           m.Seasons,
			"original_name":     m.OriginalName,
			"overview":          m.Overview,
			"poster_path":       m.PosterPath,
			"release_date":      releaseDate,
			"still_path":        "",
			"vote_average":      m.VoteAverage,
			"vote_count":        m.VoteCount,
			"source":            "Lampa",
			"original_language": m.OriginalLanguage,
			"video":             m.Video,
			"update_date":       m.UpdateDate,
			"release_quality":   qualityText,
			"create_date":       torr.CreateDate,
		})
	}

	c.JSON(200, gin.H{
		"page":          page,
		"results":       results,
		"total_pages":   totalPages,
		"total_results": totalResults,
	})
}
