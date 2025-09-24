package releases

import (
	"NUMParser/db"
	"NUMParser/db/models"
	"NUMParser/utils"
	"log"
	"sort"
)

func GetFourKMovies() {
	torrs := db.GetTorrs()
	var list []*models.TorrentDetails

	for _, torr := range torrs {
		if torr.Categories == models.CatMovie && torr.VideoQuality >= 300 {
			list = append(list, torr)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].CreateDate == list[j].CreateDate {
			if list[i].VideoQuality == list[j].VideoQuality {
				return list[i].AudioQuality > list[j].AudioQuality
			}
			return list[i].VideoQuality > list[j].VideoQuality
		}
		return list[i].CreateDate.After(list[j].CreateDate)
	})

	list = utils.UniqueTorrList(list)

	ents := FillTMDB("4K", true, list, 1000)

	log.Println("Found torrents:", len(ents))
	log.Println("All torrents:", len(list))

	save("4k_id.json", ents)

	utils.FreeOSMemGC()
}
