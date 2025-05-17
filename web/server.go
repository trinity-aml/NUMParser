package web

import (
	"NUMParser/config"
	"NUMParser/db"
	"NUMParser/db/models"
	"NUMParser/db/tmdb"
	"NUMParser/utils"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// func compareQuality(qualityI, qualityJ string) bool {
// 	qualityMap := map[string]int{
// 		"4K":    4,
// 		"1080p": 3,
// 		"720p":  2,
// 		"SD":    1,
// 	}
// 	return qualityMap[qualityI] > qualityMap[qualityJ]
// }

//func isAnimation(e *models.Entity) bool {
//	for _, g := range e.Genres {
//		if g == nil {
//			continue
//		}
//		name := strings.ToLower(g.Name)
//		if name == "мультфильм" || name == "animation" {
//			return true
//		}
//	}
//	return false
//}

//var (
//	route        *gin.Engine
//	cachedMovies struct {
//		sync.RWMutex
//		movies     []*models.Entity
//		lastUpdate time.Time
//	}
//	cacheDuration = 5 * time.Minute
//)

//func updateMoviesCache() {
//	entities := tmdb.GetAllMovies()
//	var movies []*models.Entity
//	for _, m := range entities {
//		if m.MediaType == "movie" && !isAnimation(m) {
//			categories := tmdb.GetReleaseCategoriesByTMDBID(m.ID)
//			if categories == "Movie" {
//				if m.GetTorrent() == nil {
//					m.SetTorrent(&models.TorrentDetails{})
//				}
//				movies = append(movies, m)
//			}
//		}
//	}
//
//	// Предварительно получаем все необходимые данные
//	for _, m := range movies {
//		if m.GetTorrent() == nil {
//			m.SetTorrent(&models.TorrentDetails{})
//		}
//		m.GetTorrent().CreateDate = tmdb.GetReleaseCreateDateByTMDBID(m.ID)
//		m.GetTorrent().VideoQuality = getQualityValue(tmdb.GetReleaseQualityByTMDBID(m.ID))
//	}
//
//	sort.Slice(movies, func(i, j int) bool {
//		if movies[i].GetTorrent().CreateDate.Equal(movies[j].GetTorrent().CreateDate) {
//			return movies[i].GetTorrent().VideoQuality > movies[j].GetTorrent().VideoQuality
//		}
//		return movies[i].GetTorrent().CreateDate.After(movies[j].GetTorrent().CreateDate)
//	})
//
//	cachedMovies.Lock()
//	cachedMovies.movies = movies
//	cachedMovies.lastUpdate = time.Now()
//	cachedMovies.Unlock()
//}

//func getCachedMovies() []*models.Entity {
//	cachedMovies.RLock()
//	defer cachedMovies.RUnlock()
//
//	if time.Since(cachedMovies.lastUpdate) > cacheDuration {
//		go updateMoviesCache()
//	}
//
//	return cachedMovies.movies
//}

var (
	route        *gin.Engine
	cachedMovies struct {
		sync.RWMutex
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

	// Создаём 4 отдельных списка
	var (
	//moviesNew    []*models.Entity
	//movies       []*models.Entity
	//moviesRUNew  []*models.Entity
	//moviesRU     []*models.Entity
	//tvShowsNew   []*models.Entity
	//tvShows      []*models.Entity
	//tvShowsRUNew []*models.Entity
	//tvShowsRU    []*models.Entity
	)
	//for _, m := range movieEntities {
	//
	//	torr := tmdb.GetTorrentDetailsByTMDBID(m.ID)
	//	if torr != nil {
	//		m.SetTorrent(torr)
	//	}
	//
	//	if m.GetTorrent() != nil {
	//		if torr.Categories == "Movie" {
	//			if m.OriginalLanguage == "ru" {
	//				if utils.Abs(torr.Year-time.Now().Year()) < 2 && torr.VideoQuality >= 200 {
	//					moviesRUNew = append(moviesRUNew, m)
	//				} else {
	//					moviesRU = append(moviesRU, m)
	//				}
	//			} else {
	//				if utils.Abs(torr.Year-time.Now().Year()) < 2 && torr.VideoQuality >= 200 {
	//					moviesNew = append(moviesNew, m)
	//				} else {
	//					movies = append(movies, m)
	//				}
	//			}
	//		}
	//	}
	//}
	//moviesRUNew, moviesRU := filterEntitiesByCategory(movieEntities, models.CatMovie, "ru", 2, 200)
	//moviesNew, movies := filterEntitiesByCategory(movieEntities, models.CatMovie, "", 2, 200)
	//moviesRUNew, moviesRU := filterEntitiesByCategory(movieEntities, models.CatMovie, "ru", 2, 200)
	//moviesNew, movies := filterEntitiesByCategory(movieEntities, models.CatMovie, "notru", 2, 200)
	moviesRuNew, moviesRu := filterEntitiesByCategory(movieEntities, []string{models.CatMovie}, "ru", 2, 200)
	moviesNew, movies := filterEntitiesByCategory(movieEntities, []string{models.CatMovie}, "notru", 2, 200)
	tvShowRuNew, tvShowRu := filterEntitiesByCategory(tvEntities, []string{models.CatSeries}, "ru", 2, 200)
	tvShowNew, tvShow := filterEntitiesByCategory(tvEntities, []string{models.CatSeries}, "notru", 2, 200)
	cartoonMoviesNew, cartoonMovies := filterEntitiesByCategory(movieEntities, []string{models.CatCartoonMovie}, "all", 2, 200)
	cartoonSeriesNew, cartoonSeries := filterEntitiesByCategory(tvEntities, []string{models.CatCartoonSeries}, "all", 2, 200)
	//animeNew, anime := filterEntitiesByCategory(movieEntities, []string{models.CatAnime}, "all", 2, 200)
	movies4k, movies4kNew := filterEntitiesByCategory(movieEntities, []string{models.CatMovie}, "all", 4, 300)
	sortEntities := func(entities []*models.Entity) {
		sort.Slice(entities, func(i, j int) bool {
			if entities[i].GetTorrent().CreateDate.Equal(entities[j].GetTorrent().CreateDate) {
				return entities[i].GetTorrent().VideoQuality > entities[j].GetTorrent().VideoQuality
			}
			return entities[i].GetTorrent().CreateDate.After(entities[j].GetTorrent().CreateDate)
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

	// Сортируем все списки
	sortEntities(movies4kNew)
	sortReleaseDate(movies4k)
	sortEntities(moviesNew)
	sortEntities(moviesRuNew)
	sortReleaseDate(moviesRu)
	sortReleaseDate(movies)
	//sortEntities(moviesRU)
	sortEntities(tvShowRuNew)
	sortReleaseDate(tvShowRu)
	sortEntities(tvShowNew)
	sortReleaseDate(tvShow)
	sortEntities(cartoonMoviesNew)
	sortEntities(cartoonSeriesNew)
	sortReleaseDate(cartoonMovies)
	sortReleaseDate(cartoonSeries)

	// Блокируем и обновляем кэш
	cachedMovies.Lock()
	defer cachedMovies.Unlock()

	// Здесь нужно определить как хранить списки в cachedMovies
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

// func filterEntitiesByCategory(
//
//	entities []*models.Entity,
//	category string,
//	lang string,
//	yearDelta int,
//	minQuality int,
//
//	) (newList, allList []*models.Entity) {
//		for _, m := range entities {
//			torr := tmdb.GetTorrentDetailsByTMDBID(m.ID)
//			if torr != nil {
//				m.SetTorrent(torr)
//			}
//			if m.GetTorrent() != nil && torr.Categories == category {
//				if m.OriginalLanguage == lang {
//					if utils.Abs(torr.Year-time.Now().Year()) < yearDelta && torr.VideoQuality >= minQuality {
//						newList = append(newList, m)
//					} else {
//						allList = append(allList, m)
//					}
//				} else if lang == "" { // если не фильтруем по языку
//					if utils.Abs(torr.Year-time.Now().Year()) < yearDelta && torr.VideoQuality >= minQuality {
//						newList = append(newList, m)
//					} else {
//						allList = append(allList, m)
//					}
//				}
//			}
//		}
//		return
//	}

func filterEntitiesByCategory(
	entities []*models.Entity,
	categories []string,
	langMode string, // "split", "all", "ru", "notru"
	yearDelta int,
	minQuality int,
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
					if utils.Abs(torr.Year-time.Now().Year()) < yearDelta && torr.VideoQuality >= minQuality {
						newList = append(newList, m)
					} else {
						allList = append(allList, m)
					}
				}
			}
		}
	}
	return
}

//func filterEntitiesByCategory(
//	entities []*models.Entity,
//	categories []string, // список категорий
//	lang string, // "ru" для русских, "notru" для остальных, "" для всех
//	yearDelta int,
//	minQuality int,
//) (newList, allList []*models.Entity) {
//	categorySet := make(map[string]struct{})
//	for _, cat := range categories {
//		categorySet[cat] = struct{}{}
//	}
//	for _, m := range entities {
//		torr := tmdb.GetTorrentDetailsByTMDBID(m.ID)
//		if torr != nil {
//			m.SetTorrent(torr)
//		}
//		if m.GetTorrent() != nil {
//			if _, ok := categorySet[torr.Categories]; ok {
//				isRu := m.OriginalLanguage == "ru"
//				if (lang == "ru" && isRu) || (lang == "notru" && !isRu) || (lang == "") {
//					if utils.Abs(torr.Year-time.Now().Year()) < yearDelta && torr.VideoQuality >= minQuality {
//						newList = append(newList, m)
//					} else {
//						allList = append(allList, m)
//					}
//				}
//			}
//		}
//	}
//	return
//}

//func filterEntitiesByCategory(
//	entities []*models.Entity,
//	category string,
//	lang string, // "ru" для русских, "notru" для всех кроме русских, "" для всех
//	yearDelta int,
//	minQuality int,
//) (newList, allList []*models.Entity) {
//	for _, m := range entities {
//		torr := tmdb.GetTorrentDetailsByTMDBID(m.ID)
//		if torr != nil {
//			m.SetTorrent(torr)
//		}
//		if m.GetTorrent() != nil && torr.Categories == category {
//			isRu := m.OriginalLanguage == "ru"
//			if (lang == "ru" && isRu) || (lang == "notru" && !isRu) || (lang == "") {
//				if utils.Abs(torr.Year-time.Now().Year()) < yearDelta && torr.VideoQuality >= minQuality {
//					newList = append(newList, m)
//				} else {
//					allList = append(allList, m)
//				}
//			}
//		}
//	}
//	return
//}

//func getQualityValue(quality string) int {
//	qualityMap := map[string]int{
//		"4K":    4,
//		"1080p": 3,
//		"720p":  2,
//		"SD":    1,
//	}
//	return qualityMap[quality]
//}

//var (
//	moviesNewT   []*models.Entity
//	moviesRUNewT []*models.Entity
//	moviesRUT    []*models.Entity
//	moviesT      []*models.Entity
//)

//func GetMoviess() []*models.Entity {
//	movies := tmdb.GetAllMovies()
//	for _, m := range movies {
//
//		torr := tmdb.GetTorrentDetailsByTMDBID(m.ID)
//		if torr != nil {
//			m.SetTorrent(torr)
//		}
//
//		if m.GetTorrent() != nil {
//			if torr.Categories == "Movie" {
//				//if m.GetTorrent().Categories == "Movie" {
//				if m.OriginalLanguage == "ru" {
//					if utils.Abs(torr.Year-time.Now().Year()) < 2 && torr.VideoQuality >= 200 {
//						//if utils.Abs(m.GetTorrent().Year-time.Now().Year()) < 2 && m.GetTorrent().VideoQuality > 200 {
//						moviesRUNewT = append(moviesRUNewT, m)
//					}
//					moviesRUT = append(moviesRUT, m)
//				} else {
//					if utils.Abs(torr.Year-time.Now().Year()) < 2 && torr.VideoQuality >= 200 {
//						moviesNewT = append(moviesNewT, m)
//					}
//					moviesT = append(moviesT, m)
//				}
//			}
//		}
//	}
//
//	sortEntities := func(entities []*models.Entity) {
//		sort.Slice(entities, func(i, j int) bool {
//			if entities[i].GetTorrent().CreateDate.Equal(entities[j].GetTorrent().CreateDate) {
//				return entities[i].GetTorrent().VideoQuality > entities[j].GetTorrent().VideoQuality
//			}
//			return entities[i].GetTorrent().CreateDate.After(entities[j].GetTorrent().CreateDate)
//		})
//	}
//	sortEntities(moviesNewT)
//	sortEntities(moviesRUNewT)
//	sortEntities(moviesRUT)
//	sortEntities(moviesT)
//
//	return moviesNewT
//}

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

	//// Новые фильмы тест Апи
	//r.GET("/api/lampac/movies_new_test", func(c *gin.Context) {
	//	c.Header("Access-Control-Allow-Origin", "*")
	//	c.Header("Access-Control-Allow-Headers", "Content-Type")
	//	c.Header("Content-Type", "application/json")
	//	page := getPageParam(c)
	//	movies := GetMoviess()
	//	sendMoviesResponse(c, movies, page)
	//})

	// Фильмы в высоком качестве
	r.GET("/api/lampac/4k", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.Movies4k, page)
	})

	// Новые фильмы
	r.GET("/api/lampac/movies_new", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.MoviesNew, page)
	})

	// Новые русские фильмы
	r.GET("/api/lampac/movies_ru_new", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.MoviesRuNew, page)
	})

	// Фильмы
	r.GET("/api/lampac/movies", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.Movies, page)
	})

	// Русские фильмы (без мультфильмов)
	r.GET("/api/lampac/movies_ru", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.MoviesRu, page)
	})

	//r.GET("/api/lampac/russian_movies", func(c *gin.Context) {
	//	c.Header("Access-Control-Allow-Origin", "*")
	//	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	//	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	//
	//	page := getPageParam(c)
	//	entities := tmdb.GetAllMovies()
	//	var russianMovies []*models.Entity
	//	for _, m := range entities {
	//		if m.OriginalLanguage == "ru" && m.MediaType == "movie" && !isAnimation(m) {
	//			russianMovies = append(russianMovies, m)
	//		}
	//	}
	//	sort.Slice(russianMovies, func(i, j int) bool {
	//		return russianMovies[i].UpdateDate.After(russianMovies[j].UpdateDate)
	//	})
	//	sendMoviesResponse(c, russianMovies, page)
	//})

	// Мультфильмы (только фильмы)
	r.GET("/api/lampac/cartoons", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.CartoonMoviesNew, page)
	})

	// Мультсериалы (только сериалы)
	r.GET("/api/lampac/cartoons_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.CartoonMoviesRuNew, page)
	})

	// Сериалы (без мультсериалов)
	r.GET("/api/lampac/all_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.TVShowNew, page)
	})

	// Русские сериалы (без мультсериалов)
	r.GET("/api/lampac/russian_tv", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		page := getPageParam(c)
		cached := GetCachedMovies()
		sendMoviesResponse(c, cached.TVShowRu, page)
	})

	return r
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
