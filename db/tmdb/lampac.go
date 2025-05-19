package tmdb

import (
	"NUMParser/db/db"
	"NUMParser/db/models"
	"NUMParser/db/utils"
	"bytes"
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"sort"
)

func GetAllMovies() []*models.Entity {
	var movies []*models.Entity
	db.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Movies"))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(_, v []byte) error {
			var e *models.Entity
			if err := json.Unmarshal(v, &e); err == nil && e != nil {
				movies = append(movies, e)
			}
			return nil
		})
	})
	return movies
}

func GetAllTV() []*models.Entity {
	var tvs []*models.Entity
	db.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("TV"))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(_, v []byte) error {
			var e *models.Entity
			if err := json.Unmarshal(v, &e); err == nil && e != nil {
				tvs = append(tvs, e)
			}
			return nil
		})
	})
	return tvs
}

// Получить все данные торрента по TMDB ID
// func GetTorrentDetailsByTMDBID(tmdbID int64) *models.TorrentDetails {
// 	var key string
// 	db.DB.View(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte("TMDB"))
// 		if bucket == nil {
// 			return nil
// 		}
// 		bucket = bucket.Bucket([]byte("Index"))
// 		if bucket == nil {
// 			return nil
// 		}
// 		c := bucket.Cursor()
// 		tmdbIDBytes := utils.I2B(tmdbID)
// 		for k, v := c.First(); k != nil; k, v = c.Next() {
// 			if bytes.Equal(v, tmdbIDBytes) {
// 				key = string(k)
// 				break
// 			}
// 		}
// 		return nil
// 	})
// 	if key == "" {
// 		return nil
// 	}

// 	var details *models.TorrentDetails
// 	db.DB.View(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte("Rutor"))
// 		if bucket == nil {
// 			return nil
// 		}
// 		bucket = bucket.Bucket([]byte("Torrents"))
// 		if bucket == nil {
// 			return nil
// 		}
// 		v := bucket.Get([]byte(key))
// 		if v == nil {
// 			return nil
// 		}
// 		var t models.TorrentDetails
// 		if err := json.Unmarshal(v, &t); err == nil {
// 			details = &t
// 		}
// 		return nil
// 	})

// 	return details
// }

// Получить лучший торрент по TMDB ID с учетом категории и сортировки
func GetTorrentDetailsByTMDBID(tmdbID int64) *models.TorrentDetails {
	var torrents []*models.TorrentDetails

	// Получаем все ключи торрентов для данного TMDB ID
	db.DB.View(func(tx *bolt.Tx) error {
		// Получаем все ключи из TMDB -> Index
		tmdbBucket := tx.Bucket([]byte("TMDB"))
		if tmdbBucket == nil {
			return nil
		}
		indexBucket := tmdbBucket.Bucket([]byte("Index"))
		if indexBucket == nil {
			return nil
		}

		tmdbIDBytes := utils.I2B(tmdbID)
		c := indexBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if bytes.Equal(v, tmdbIDBytes) {
				key := string(k)

				// Получаем сам торрент
				rutorBucket := tx.Bucket([]byte("Rutor"))
				if rutorBucket == nil {
					continue
				}
				torrentsBucket := rutorBucket.Bucket([]byte("Torrents"))
				if torrentsBucket == nil {
					continue
				}

				v := torrentsBucket.Get([]byte(key))
				if v == nil {
					continue
				}

				var t models.TorrentDetails
				if err := json.Unmarshal(v, &t); err == nil {
					torrents = append(torrents, &t)
				}
			}
		}
		return nil
	})

	if len(torrents) == 0 {
		return nil
	}

	// Сортируем по приоритету: дата (свежие), качество видео, качество аудио
	sort.Slice(torrents, func(i, j int) bool {
		if torrents[i].CreateDate.Equal(torrents[j].CreateDate) {
			if torrents[i].VideoQuality == torrents[j].VideoQuality {
				return torrents[i].AudioQuality > torrents[j].AudioQuality
			}
			return torrents[i].VideoQuality > torrents[j].VideoQuality
		}
		return torrents[i].CreateDate.After(torrents[j].CreateDate)
	})

	return torrents[0] // Возвращаем лучший торрент
}
