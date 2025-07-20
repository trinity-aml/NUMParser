package releases

import (
	"NUMParser/config"
	"NUMParser/db"
	"NUMParser/db/models"
	"NUMParser/ml"
	"NUMParser/movies/tmdb"
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
)

var (
	colls = []*ml.Collection{
		{"Наши друзья", "Семейные фильмы про животных", "Фильмы про кошек, собак и других животных друзей людей, начиная с 1990 года и жанрами комедия, семейный фильм"},
		{"Космическая одиссея", "Классика фантастики о покорении космоса", "Научно-фантастические фильмы про освоение космоса до 2000 года"},
		{"Вселенная Нолана", "Хронология фильмов культового режиссера", "Фильмы Кристофера Нолана в хронологическом порядке"},
		{"Ностальгические 90-е", "Знаковые подростковые комедии эпохи", "Популярные подростковые комедии 1995-1999 годов"},
		{"Винтажные ужасы", "Культовые хорроры с практическими эффектами", "Культовые фильмы ужасов 1970-1980 годов с практическими эффектами"},
		{"Киберпанк будущего", "Цифровые антиутопии о технологиях", "Футуристичные антиутопии про хакеров и искусственный интеллект 2010-2023 годов"},
		{"Мистика старых замков", "Жуткие истории в готических интерьерах", "Фильмы ужасов и триллеры, действие которых происходит в старых замках и таинственных особняках"},
		{"Эпические сражения", "Грандиозные баталии в истории и фэнтези", "Исторические фильмы и фэнтези с масштабными битвами и военными кампаниями"},
		{"Герои в масках", "Эволюция супергеройского жанра", "Фильмы о супергероях, спасающих мир, от классики до современных блокбастеров"},
		{"Загадки космоса", "Тайны Вселенной в документальном кино", "Документальные и научно-популярные фильмы о тайнах Вселенной, чёрных дырах и внеземной жизни"},
		{"Кулинарное путешествие", "Гастрономические истории для гурманов", "Фильмы, посвящённые еде, кулинарии и ресторанному бизнесу, вдохновляющие на гастрономические приключения"},
		{"Режиссёрский дебют", "Первые шаги великих кинематографистов", "Первые полнометражные фильмы известных режиссёров, раскрывающие их уникальный стиль"},
		{"Сны наяву", "Сюрреалистические исследования сознания", "Сюрреалистические и артхаусные фильмы, исследующие границы реальности и сновидений"},
		{"Один актёр, много лиц", "Трансформации и многогранность талантов", "Фильмы, где один актёр исполняет несколько ролей или сильно меняется на протяжении карьеры"},
		{"Кино, изменившее мир", "Фильмы-революционеры в истории кино", "Фильмы, оказавшие значительное влияние на культуру, общество или кинематограф"},
		{"Уютные вечера", "Тёплые фильмы для семейного просмотра", "Добрые и светлые фильмы для семейного просмотра или отдыха после долгого дня"},
		{"Вдохновляющие истории", "Подлинные истории преодоления", "Фильмы, основанные на реальных событиях, о людях, которые преодолели трудности и достигли успеха"},
		{"Приключения для всей семьи", "Захватывающие истории для всех возрастов", "Приключенческие фильмы и мультфильмы, которые понравятся как детям, так и взрослым"},
		{"Детективы старой школы", "Классические расследования с харизматичными сыщиками", "Классические детективные фильмы до 1970 года с харизматичными сыщиками и непредсказуемыми развязками"},
		{"Безумные учёные", "Гениальные изобретатели и их эксперименты", "Фильмы про гениальных, но эксцентричных учёных и их изобретения, приведшие к неожиданным последствиям"},
		{"Спорт в кино", "Драмы и комедии о спортивных достижениях", "Драмы и комедии о спорте, победах, поражениях и силе духа спортсменов"},
		{"Городские легенды", "Ужасы, основанные на современных мифах", "Фильмы ужасов и триллеры, основанные на популярных городских легендах и мифах"},
		{"Музыкальные биопики", "Жизнь великих музыкантов на экране", "Биографические фильмы о жизни и творчестве известных музыкантов и исполнителей"},
		{"Постапокалиптический мир", "Борьба за выживание после катастрофы", "Фильмы о жизни после глобальной катастрофы, выживании и новом обществе"},
		{"Мир дикой природы", "Документальные путешествия по заповедным местам", "Документальные фильмы о флоре и фауне нашей планеты, красоте и суровости дикой природы"},
		{"Французская новая волна", "Авангардные фильмы, изменившие кино", "Знаковые фильмы французской Новой волны, изменившие представление о кинематографе"},
		{"Психологические триллеры", "Исследование тёмных уголков разума", "Фильмы с напряжённым сюжетом, исследующие тёмные стороны человеческой психики"},
		{"Путешествия во времени", "Парадоксы и альтернативные реальности", "Фильмы, где герои перемещаются в прошлое или будущее, меняя ход истории"},
		{"Искусственный интеллект", "Фильмы о будущем разумных машин", "Фильмы о разумных машинах, их развитии и взаимодействии с человечеством"},
		{"Битвы роботов", "Фантастические сражения механических гигантов", "Фантастические фильмы про гигантских роботов и их сражения"},
		{"Экранизации Шекспира", "Вечная классика в кинематографе", "Фильмы, основанные на произведениях Уильяма Шекспира, от классических до современных адаптаций"},
		{"Культовые комедии 80-х", "Легендарный юмор эпохи диско", "Легендарные комедии 1980-х годов, которые до сих пор смешат зрителей"},
		{"Неизвестные планеты", "Фантастические миры и инопланетные цивилизации", "Фантастические фильмы о приключениях на далёких инопланетных мирах"},
		{"Фильмы о хакерах", "Кибертриллеры о цифровой эпохе", "Технотриллеры и драмы о взломах, киберпреступности и цифровом мире"},
		{"Закулисье Голливуда", "Тайны кинематографической индустрии", "Фильмы, рассказывающие о создании кино, жизни актёров и режиссёров"},
		{"Выживание в экстремальных условиях", "Борьба человека со стихией", "Фильмы о людях, столкнувшихся с природными катаклизмами и борющихся за жизнь"},
		{"Вампиры и оборотни", "Вечная борьба мистических существ", "Фильмы о мистических существах, их противостоянии и историях любви"},
		{"Фильмы-катастрофы", "Масштабные бедствия и человеческий героизм", "Эпические фильмы о природных и техногенных катастрофах и героизме людей"},
		{"Мистические детективы", "Расследования с элементами сверхъестественного", "Детективы с элементами мистики, где разгадка кроется в сверхъестественном"},
		{"Антиутопии будущего", "Мрачные картины тоталитарного завтра", "Фильмы, изображающие мрачное будущее человечества под тоталитарным контролем"},
		{"Фильмы про тюрьмы", "Драмы о жизни за решёткой и побегах", "Драмы и триллеры о жизни в заключении, побегах и борьбе за свободу"},
		{"Магия и волшебство", "Фэнтези о чародеях и волшебных мирах", "Фэнтези фильмы о чародеях, магических мирах и приключениях"},
		{"Психологические драмы", "Глубокие исследования человеческих отношений", "Глубокие фильмы, исследующие человеческие отношения, эмоции и конфликты"},
		{"Путешествия по миру", "Приключения в экзотических локациях", "Приключенческие фильмы, действие которых происходит в экзотических странах и неизведанных местах"},
		{"Классические нуары", "Чёрно-белые детективы с роковыми женщинами", "Черно-белые фильмы-нуар, с роковыми женщинами, циничными детективами и мрачной атмосферой"},
		{"Фильмы про мафию", "Гангстерские саги о власти и предательстве", "Гангстерские саги о криминальном мире, власти и предательстве"},
		{"Экранизации комиксов", "Киноадаптации графических романов", "Фильмы, основанные на популярных комиксах, не относящиеся к супергероике"},
		{"Азиатские боевики", "Энергичные единоборства и динамичные сюжеты", "Динамичные боевики из Азии с захватывающими единоборствами и оригинальными сюжетами"},
		{"Военные драмы", "Фильмы о цене войны и человеческих судьбах", "Фильмы о войне, её последствиях и человеческих судьбах"},
		{"Биографии великих", "Жизнеописания выдающихся исторических фигур", "Биографические фильмы о выдающихся личностях в истории, науке и искусстве"},
	}
)

func GetCollections() {
	log.Println("Search collections")
	log.Println("Collections count:", len(colls))

	var collsId []*CollectionId

	for c, coll := range colls {
		fmt.Println("Get movies for coll:", coll.Title, c+1, "/", len(colls))
		fmt.Println("Overview:", coll.Overview)
		fmt.Println("Prompt:", coll.Prompt)

		movies, err := ml.SendGeminiRequest(coll)
		if err != nil {
			log.Printf("Gemini error for '%s': %v", coll.Title, err)
			continue
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
				collsId = append(collsId, getCollectionId(coll, found))
			}
		}

		fmt.Println()
	}

	//Save collections

	os.MkdirAll(config.SaveReleasePath, 0777)
	fname := filepath.Join(config.SaveReleasePath, "collections_id.json")

	ff, err := os.Create(fname)
	if err != nil {
		return
	}
	defer ff.Close()
	//zw := gzip.NewWriter(ff)
	//defer zw.Close()

	err = json.NewEncoder(ff).Encode(collsId)
	if err != nil {
		log.Println("Error save collections:", err)
	}

	ff, err = os.Create(fname + ".z")
	if err != nil {
		return
	}
	defer ff.Close()
	zw := gzip.NewWriter(ff)
	defer zw.Close()

	err = json.NewEncoder(zw).Encode(collsId)

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
