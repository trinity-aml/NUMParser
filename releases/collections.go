package releases

import (
	"NUMParser/config"
	"NUMParser/db"
	"NUMParser/db/models"
	"NUMParser/ml"
	"NUMParser/movies/tmdb"
	"NUMParser/releases/cover"
	"NUMParser/utils"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func GetCollections() {
	log.Println("Update collections")

	var collections []*ml.Collection
	var err error
	for {
		collections, err = ml.GenCollection(1)
		if err != nil {
			log.Println("Error while generating collection:", err)
			return
		}
		if collections != nil && len(collections) > 0 && collections[0].Title != "" && collections[0].Overview != "" && collections[0].Prompt != "" {
			break
		}
		time.Sleep(time.Second)
		log.Println("Gen coll is empty, try again")
	}

	ml.CollsConfig.Collections = ml.CollsConfig.Collections[:len(ml.CollsConfig.Collections)-1]
	ml.CollsConfig.Collections = append(collections[:1], ml.CollsConfig.Collections...)

	log.Println("Collections count:", len(ml.CollsConfig.Collections))

	var collsId []*CollectionId

	os.RemoveAll(filepath.Join(config.SaveReleasePath, "imgs"))
	os.MkdirAll(filepath.Join(config.SaveReleasePath, "imgs"), 0777)

	for c, coll := range ml.CollsConfig.Collections {
		fmt.Println("Get movies for coll:", coll.Title, c+1, "/", len(ml.CollsConfig.Collections))
		fmt.Println("Overview:", coll.Overview)
		fmt.Println("Prompt:", coll.Prompt)

		var movies []*ml.MovieInfo
		for i := 0; i < 10; i++ {
			movies, err = ml.GetCollectionMovies(coll)
			if err == nil {
				break
			}
			log.Printf("Gemini error for '%s': %v", coll.Title, err)
			time.Sleep(time.Second * 5)
		}

		var ents []*models.Entity
		var found []*models.Entity

		fmt.Println("Search tmdb ids:")
		for _, movie := range movies {
			fmt.Printf("%s (%s) == ", movie.Title, movie.Year)
			e := findTmdbMlMovie(movie)
			if e != nil {
				fmt.Printf("%s | %s (%v): %v, %.2f\n", e.Title, e.OriginalTitle, e.Year, e.ID, e.VoteAverage*math.Log(float64(e.VoteCount)))
				ents = append(ents, e)
			} else {
				fmt.Println("TMDB not found")
			}
		}

		fmt.Println("Found tmdb for coll:", coll.Title, len(ents))
		if len(ents) > 0 {
			fmt.Println("Search torrents for cols")
			for i, e := range ents {
				list := db.SearchTorr(e.Title + " " + e.OriginalTitle + " " + e.Year)
				if len(list) == 0 {
					names := getEnNames(e.Titles)
					for _, name := range names {
						list = db.SearchTorr(e.Title + " " + name + " " + e.Year)
						if len(list) > 0 {
							break
						}
					}
				}
				if len(list) > 0 {
					sort.Slice(list, func(i, j int) bool {
						if list[i].CreateDate == list[j].CreateDate {
							if list[i].VideoQuality == list[j].VideoQuality {
								return list[i].AudioQuality > list[j].AudioQuality
							}
							return list[i].VideoQuality > list[j].VideoQuality
						}
						return list[i].CreateDate.After(list[j].CreateDate)
					})

					e.SetTorrent(list[0])
					found = append(found, e)
				}
				log.Println("Fill coll:", coll.Title, i+1, "/", len(ents))
				utils.FreeOSMemGC()
			}

			if len(found) > 0 {
				fmt.Println("Found torrents for coll:", coll.Title, len(found))
				collId := getCollectionId(coll, found)
				collsId = append(collsId, collId)

				// строим картинку
				var urls []string
				var backdrop string

				for _, e := range found {
					if e.PosterPath != "" {
						urls = append(urls, strings.ReplaceAll(e.PosterPath, "w342", "w500"))
						if backdrop == "" {
							backdrop = strings.ReplaceAll(e.BackdropPath, "w780", "w1280")
						}
					}
					if len(urls) == 4 {
						break
					}
				}

				if backdrop == "" {
					for _, e := range found {
						if e.BackdropPath != "" {
							backdrop = strings.ReplaceAll(e.BackdropPath, "w780", "w1280")
							break
						}
					}
				}

				// если постеров нет вообще — просто не вызываем BuildCover
				if len(urls) > 0 {
					imgfn := strconv.Itoa(c) + ".png"
					imgfn = filepath.Join(config.SaveReleasePath, "imgs", imgfn)

					for e := 0; e < 10; e++ {
						err = cover.BuildCover(urls, backdrop, imgfn)
						if err == nil {
							break
						}
					}
					if err != nil {
						log.Println("Cover build failed:", err)
						imgfn = ""
					}
					collId.PosterPath = strings.ReplaceAll(imgfn, "public", "")
				}
			}
		}

		fmt.Println()
	}

	ml.SaveConfig()

	//Save collections

	os.MkdirAll(config.SaveReleasePath, 0777)
	fname := filepath.Join(config.SaveReleasePath, "colls_movie_ai.json")

	ff, err := os.Create(fname)
	if err != nil {
		return
	}
	defer ff.Close()
	zw := gzip.NewWriter(ff)
	defer zw.Close()

	err = json.NewEncoder(zw).Encode(collsId)
	if err != nil {
		log.Println("Error save collections:", err)
	}

	utils.FreeOSMemGC()
}

func getCollectionId(coll *ml.Collection, ents []*models.Entity) *CollectionId {
	rid := &CollectionId{
		Name:         coll.Title,
		Overview:     coll.Overview,
		Parts:        nil,
		PosterPath:   "",
		BackdropPath: "",
	}

	for _, e := range ents {
		if e != nil && e.GetTorrent() != nil {
			var countries []string
			if len(e.ProductionCountries) > 0 {
				for _, c := range e.ProductionCountries {
					countries = append(countries, c.Iso31661)
				}
			} else {
				countries = e.OriginCountry
			}
			d := e.GetTorrent()
			t := Torrent{
				Name:     d.Name,
				Date:     d.CreateDate.Format("02.01.2006"),
				Magnet:   d.Magnet,
				Size:     d.Size,
				Upload:   strconv.Itoa(d.Seed),
				Download: strconv.Itoa(d.Peer),
				Source:   "Rutor",
				Link:     d.Link,
				Quality:  d.VideoQuality,
				Voice:    d.AudioQuality,
			}
			tid := &TmdbId{
				Id:          e.ID,
				MediaType:   e.MediaType,
				GenreIds:    e.GenresIds,
				VoteAverage: e.VoteAverage,
				VoteCount:   e.VoteCount,
				Countries:   countries,
				ReleaseDate: e.ReleaseDate,
				Torrent:     []*Torrent{&t},
			}
			rid.Parts = append(rid.Parts, tid)
		}
	}

	sort.Slice(rid.Parts, func(i, j int) bool {
		rankI := rid.Parts[i].VoteAverage * math.Log(float64(rid.Parts[i].VoteCount))
		rankJ := rid.Parts[j].VoteAverage * math.Log(float64(rid.Parts[j].VoteCount))
		return rankI > rankJ
	})

	sort.Slice(ents, func(i, j int) bool {
		rankI := ents[i].VoteAverage * math.Log(float64(ents[i].VoteCount))
		rankJ := ents[j].VoteAverage * math.Log(float64(ents[j].VoteCount))
		return rankI > rankJ
	})

	return rid
}

func findTmdbMlMovie(movie *ml.MovieInfo) *models.Entity {
	list := tmdb.Search(true, movie.Title)
	yearml, _ := strconv.Atoi(movie.Year)

	list = utils.Filter(list, func(i int, e *models.Entity) bool {
		if len(e.ReleaseDate) > 6 {
			year, _ := strconv.Atoi(e.ReleaseDate[6:])
			return utils.Abs(year-yearml) > 1
		}
		return true
	})

	if len(list) == 1 {
		return list[0]
	}

	list = utils.Distinct(list, func(e *models.Entity) string {
		return strconv.FormatInt(e.ID, 10)
	})

	if len(list) == 1 {
		return list[0]
	}

	if len(list) > 0 {
		list = utils.Filter(list, func(i int, e *models.Entity) bool {
			if len(e.ReleaseDate) > 6 {
				year, _ := strconv.Atoi(e.ReleaseDate[6:])
				return utils.Abs(year-yearml) > 1
			}
			return true
		})

		list = utils.Filter(list, func(i int, e *models.Entity) bool {
			if utils.ClearStr(e.Title) == utils.ClearStr(e.OriginalTitle) &&
				utils.ClearStr(movie.Title) == utils.ClearStr(e.Title) {
				return false
			} else if utils.ClearStr(e.OriginalTitle) == utils.ClearStr(movie.Title) ||
				utils.ClearStr(e.Title) == utils.ClearStr(movie.Title) {
				return false
			}
			return true
		})
		if len(list) == 1 {
			return list[0]
		}
	}
	return nil
}
