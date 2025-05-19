package web

import (
	"NUMParser/db/models"
	"NUMParser/db/tmdb"
	"NUMParser/utils"
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
		movies4kNew        []*models.Entity
		movies4k           []*models.Entity
		moviesNew          []*models.Entity
		movies             []*models.Entity
		moviesRuNew        []*models.Entity
		moviesRu           []*models.Entity
		tvShowNew          []*models.Entity
		tvShow             []*models.Entity
		tvShowRuNew        []*models.Entity
		tvShowRu           []*models.Entity
		cartoonMovies      []*models.Entity
		cartoonMoviesNew   []*models.Entity
		cartoonMoviesRu    []*models.Entity
		cartoonMoviesRuNew []*models.Entity
		lastUpdate         time.Time
	}
	cacheDuration = 5 * time.Minute
)

func updateMoviesCache() {

	// Получаем все данные
	movieEntities := tmdb.GetAllMovies()
	tvEntities := tmdb.GetAllTV()

	// Фильтруем данные по категориям
	moviesRuNew, moviesRu := filterEntitiesByCategory(movieEntities, []string{models.CatMovie}, "ru", 2, 200, false)
	moviesNew, movies := filterEntitiesByCategory(movieEntities, []string{models.CatMovie}, "notru", 2, 200, false)
	tvShowRuNew, tvShowRu := filterEntitiesByCategory(tvEntities, []string{models.CatSeries}, "ru", 2, 200, false)
	tvShowNew, tvShow := filterEntitiesByCategory(tvEntities, []string{models.CatSeries}, "notru", 2, 200, false)
	cartoonMoviesNew, cartoonMovies := filterEntitiesByCategory(movieEntities, []string{models.CatCartoonMovie}, "all", 2, 200, false)
	cartoonSeriesNew, cartoonSeries := filterEntitiesByCategory(tvEntities, []string{models.CatCartoonSeries}, "all", 2, 200, false)
	//animeNew, anime := filterEntitiesByCategory(movieEntities, []string{models.CatAnime}, "all", 2, 200, false)
	movies4kNew, movies4k := filterEntitiesByCategory(movieEntities, []string{models.CatMovie}, "all", 4, 300, true)
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
		// sort.Slice(entities, func(i, j int) bool {
		// 	if entities[i].GetTorrent().CreateDate.Equal(entities[j].GetTorrent().CreateDate) {
		// 		return entities[i].GetTorrent().VideoQuality > entities[j].GetTorrent().VideoQuality
		// 	}
		// 	return entities[i].GetTorrent().CreateDate.After(entities[j].GetTorrent().CreateDate)
		// })
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
	cachedMovies.tvShowNew = tvShowRuNew
	cachedMovies.tvShow = tvShow
	cachedMovies.tvShowNew = tvShowNew
	cachedMovies.cartoonMoviesNew = cartoonMoviesNew
	cachedMovies.cartoonMovies = cartoonMovies
	cachedMovies.cartoonMoviesRuNew = cartoonSeriesNew
	cachedMovies.cartoonMoviesRu = cartoonSeries
	cachedMovies.lastUpdate = time.Now()
}

// Добавляем тип для возврата всех категорий
type CachedMoviesResponse struct {
	Movies4k           []*models.Entity
	Movies4kNew        []*models.Entity
	MoviesNew          []*models.Entity
	Movies             []*models.Entity
	MoviesRuNew        []*models.Entity
	MoviesRu           []*models.Entity
	TVShowNew          []*models.Entity
	TVShow             []*models.Entity
	TVShowRuNew        []*models.Entity
	TVShowRu           []*models.Entity
	CartoonMovies      []*models.Entity
	CartoonMoviesRu    []*models.Entity
	CartoonMoviesNew   []*models.Entity
	CartoonMoviesRuNew []*models.Entity
	LastUpdate         time.Time
}

// Получаем все категории фильмов
func GetCachedMovies() CachedMoviesResponse {
	cachedMovies.RLock()
	defer cachedMovies.RUnlock()

	// Если кэш устарел, обновляем в фоне
	if time.Since(cachedMovies.lastUpdate) > cacheDuration {
		go updateMoviesCache()
	}

	return CachedMoviesResponse{
		Movies4k:           cachedMovies.movies4k,
		Movies4kNew:        cachedMovies.movies4kNew,
		MoviesNew:          cachedMovies.moviesNew,
		Movies:             cachedMovies.movies,
		MoviesRuNew:        cachedMovies.moviesRuNew,
		MoviesRu:           cachedMovies.moviesRu,
		TVShowNew:          cachedMovies.tvShowNew,
		TVShow:             cachedMovies.tvShow,
		TVShowRuNew:        cachedMovies.tvShowRuNew,
		TVShowRu:           cachedMovies.tvShowRu,
		CartoonMovies:      cachedMovies.cartoonMovies,
		CartoonMoviesRu:    cachedMovies.cartoonMoviesRu,
		CartoonMoviesNew:   cachedMovies.cartoonMoviesNew,
		CartoonMoviesRuNew: cachedMovies.cartoonMoviesRuNew,
		LastUpdate:         cachedMovies.lastUpdate,
	}
}

func filterEntitiesByCategory(
	entities []*models.Entity,
	categories []string,
	langMode string, // "split", "all", "ru", "notru"
	yearDelta int,
	minQuality int,
	is4KFilter bool, // true - фильтруем 4K, false - обычные фильмы
) (newList, allList []*models.Entity) {
	categorySet := make(map[string]struct{})
	for _, cat := range categories {
		categorySet[cat] = struct{}{}
	}
	for _, m := range entities {
		torr := tmdb.GetTorrentDetailsByTMDBID(m.ID)
		if torr != nil {
			m.SetTorrent(torr)
		}
		if m.GetTorrent() != nil {
			if _, ok := categorySet[torr.Categories]; ok {
				isRu := m.OriginalLanguage == "ru"
				// Логика по языку
				use := false
				switch langMode {
				case "split":
					// split не используется напрямую, см. ниже
				case "all":
					use = true
				case "ru":
					use = isRu
				case "notru":
					use = !isRu
				}
				if use || langMode == "all" {
					is4K := torr.VideoQuality >= 300

					// Если запросили 4K - обрабатываем только 4K
					if is4KFilter && is4K {
						if utils.Abs(torr.Year-time.Now().Year()) < yearDelta {
							newList = append(newList, m) // movies4kNew
						} else {
							allList = append(allList, m) // movies4k
						}
					}

					// Если запросили обычные - обрабатываем только обычные
					if !is4KFilter && !is4K {
						if utils.Abs(torr.Year-time.Now().Year()) < yearDelta && torr.VideoQuality >= minQuality {
							newList = append(newList, m) // moviesNew
						} else if torr.VideoQuality >= minQuality {
							allList = append(allList, m) // movies
						}
					}
				}
			}
		}
	}
	return
}

func InitLampacRoutes(r *gin.RouterGroup) {

	// Инициализация кэша при старте
	updateMoviesCache()

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
		sendMoviesResponse(c, cached.CartoonMoviesNew, page)
	})

	// Мультсериалы (только сериалы)
	r.GET("/cartoons_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.CartoonMoviesRuNew, page)
	})

	// Сериалы (без мультсериалов)
	r.GET("/all_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.TVShowNew, page)
	})

	// Русские сериалы (без мультсериалов)
	r.GET("/russian_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.TVShowRu, page)
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
			"backdrop_path": m.BackdropPath,
			//"first_air_date":    first_air_date,
			"id": m.ID,
			//"name":              name,
			"name":              m.Title,
			"number_of_seasons": m.NumberOfSeasons,
			"seasons":           m.Seasons,
			//"original_name":     original_name,
			"original_name": m.OriginalName,
			"overview":      m.Overview,
			"poster_path":   m.PosterPath,
			"release_date":  releaseDate,
			//"release_date":      m.ReleaseDate,
			"still_path":        "",
			"vote_average":      m.VoteAverage,
			"vote_count":        m.VoteCount,
			"source":            "Lampa",
			"original_language": m.OriginalLanguage,
			"media_type":        m.MediaType,
			"video":             m.Video,
			"update_date":       m.UpdateDate,
			"release_quality":   qualityText,
			"categories":        torr.Categories,
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
