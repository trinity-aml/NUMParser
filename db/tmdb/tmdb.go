package tmdb

import (
	"NUMParser/db/db"
	"NUMParser/db/models"
	"NUMParser/db/utils"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"time"

	bolt "go.etcd.io/bbolt"
)

func GetMovie(id int64) *models.Entity {
	var ent *models.Entity
	db.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("TMDB"))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte("Ents"))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte("Movies"))
		if bucket == nil {
			return nil
		}
		buf := bucket.Get(utils.I2B(id))
		if len(buf) == 0 {
			return nil
		}
		return json.Unmarshal(buf, &ent)
	})
	return ent
}

func GetTV(id int64) *models.Entity {
	var ent *models.Entity
	db.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("TMDB"))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte("Ents"))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte("TV"))
		if bucket == nil {
			return nil
		}
		buf := bucket.Get(utils.I2B(id))
		if len(buf) == 0 {
			return nil
		}
		return json.Unmarshal(buf, &ent)
	})
	return ent
}

func FindIMDB(id string) *models.Entity {
	var ent *models.Entity
	db.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("TMDB"))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte("Ents"))
		if bucket == nil {
			return nil
		}
		bckt := bucket.Bucket([]byte("TV"))
		if bckt == nil {
			return nil
		}
		bckt.ForEach(func(_, v []byte) error {
			var e *models.Entity
			err := json.Unmarshal(v, &e)
			if err != nil {
				log.Fatalln("Error read from db TMDB:", err)
			}
			if e.ImdbID == id {
				ent = e
				return errors.New("")
			}
			return nil
		})
		bckt = bucket.Bucket([]byte("Movies"))
		if bckt == nil {
			return nil
		}
		bckt.ForEach(func(_, v []byte) error {
			var e *models.Entity
			err := json.Unmarshal(v, &e)
			if err != nil {
				log.Fatalln("Error read from db TMDB:", err)
			}
			if e.ImdbID == id {
				ent = e
				return errors.New("")
			}
			return nil
		})
		return nil
	})
	return ent
}

func AddTMDB(t *models.Entity) {
	t.UpdateDate = time.Now()
	db.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("TMDB"))
		if err != nil {
			return err
		}
		bucket, err = tx.CreateBucketIfNotExists([]byte("Ents"))
		if err != nil {
			return err
		}
		if t.MediaType == "movie" {
			bucket, err = tx.CreateBucketIfNotExists([]byte("Movies"))
		} else {
			bucket, err = tx.CreateBucketIfNotExists([]byte("TV"))
		}
		if err != nil {
			return err
		}
		buf, err := json.Marshal(t)
		if err != nil {
			return err
		}
		return bucket.Put(utils.I2B(t.ID), buf)
	})
}

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

// GetAllMoviesRu возвращает только фильмы на русском языке
// func GetAllMoviesRu() []*models.Entity {
// 	var movies []*models.Entity
// 	db.DB.View(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte("Movies"))
// 		if bucket == nil {
// 			return nil
// 		}
// 		return bucket.ForEach(func(_, v []byte) error {
// 			var e *models.Entity
// 			if err := json.Unmarshal(v, &e); err == nil && e != nil && e.OriginalLanguage == "ru" {
// 				movies = append(movies, e)
// 			}
// 			return nil
// 		})
// 	})
// 	return movies
// }

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

//func GetReleaseQualityByTMDBID(tmdbID int64) string {
//	var key string
//	db.DB.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("TMDB"))
//		if bucket == nil {
//			return nil
//		}
//		bucket = bucket.Bucket([]byte("Index"))
//		if bucket == nil {
//			return nil
//		}
//		c := bucket.Cursor()
//		tmdbIDBytes := utils.I2B(tmdbID)
//		for k, v := c.First(); k != nil; k, v = c.Next() {
//			if bytes.Equal(v, tmdbIDBytes) {
//				key = string(k)
//				break
//			}
//		}
//		return nil
//	})
//	if key == "" {
//		return ""
//	}
//
//	var videoQuality int
//	db.DB.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("Rutor"))
//		if bucket == nil {
//			return nil
//		}
//		bucket = bucket.Bucket([]byte("Torrents"))
//		if bucket == nil {
//			return nil
//		}
//		v := bucket.Get([]byte(key))
//		if v == nil {
//			return nil
//		}
//		var t struct {
//			VideoQuality int `json:"VideoQuality"`
//		}
//		if err := json.Unmarshal(v, &t); err == nil {
//			videoQuality = t.VideoQuality
//		}
//		return nil
//	})
//
//	switch {
//	case videoQuality >= 300:
//		return "4K"
//	case videoQuality >= 200:
//		return "1080p"
//	case videoQuality >= 100:
//		return "720p"
//	default:
//		return "SD"
//	}
//}

// Получить все данные торрента по TMDB ID
func GetTorrentDetailsByTMDBID(tmdbID int64) *models.TorrentDetails {
	var key string
	db.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("TMDB"))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte("Index"))
		if bucket == nil {
			return nil
		}
		c := bucket.Cursor()
		tmdbIDBytes := utils.I2B(tmdbID)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if bytes.Equal(v, tmdbIDBytes) {
				key = string(k)
				break
			}
		}
		return nil
	})
	if key == "" {
		return nil
	}

	var details *models.TorrentDetails
	db.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Rutor"))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte("Torrents"))
		if bucket == nil {
			return nil
		}
		v := bucket.Get([]byte(key))
		if v == nil {
			return nil
		}
		var t models.TorrentDetails
		if err := json.Unmarshal(v, &t); err == nil {
			details = &t
		}
		return nil
	})

	return details
}

//func GetReleaseCategoriesByTMDBID(tmdbID int64) string {
//	var key string
//	db.DB.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("TMDB"))
//		if bucket == nil {
//			return nil
//		}
//		bucket = bucket.Bucket([]byte("Index"))
//		if bucket == nil {
//			return nil
//		}
//		c := bucket.Cursor()
//		tmdbIDBytes := utils.I2B(tmdbID)
//		for k, v := c.First(); k != nil; k, v = c.Next() {
//			if bytes.Equal(v, tmdbIDBytes) {
//				key = string(k)
//				break
//			}
//		}
//		return nil
//	})
//	if key == "" {
//		return ""
//	}
//
//	var categories string
//	db.DB.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("Rutor"))
//		if bucket == nil {
//			return nil
//		}
//		bucket = bucket.Bucket([]byte("Torrents"))
//		if bucket == nil {
//			return nil
//		}
//		v := bucket.Get([]byte(key))
//		if v == nil {
//			return nil
//		}
//		var t struct {
//			Categories string `json:"Categories"`
//		}
//		if err := json.Unmarshal(v, &t); err == nil {
//			categories = t.Categories
//		}
//		return nil
//	})
//
//	return categories
//}
//
//func GetReleaseCreateDateByTMDBID(tmdbID int64) time.Time {
//	var key string
//	db.DB.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("TMDB"))
//		if bucket == nil {
//			return nil
//		}
//		bucket = bucket.Bucket([]byte("Index"))
//		if bucket == nil {
//			return nil
//		}
//		c := bucket.Cursor()
//		tmdbIDBytes := utils.I2B(tmdbID)
//		for k, v := c.First(); k != nil; k, v = c.Next() {
//			if bytes.Equal(v, tmdbIDBytes) {
//				key = string(k)
//				break
//			}
//		}
//		return nil
//	})
//	if key == "" {
//		return time.Time{}
//	}
//
//	var createDate time.Time
//	db.DB.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("Rutor"))
//		if bucket == nil {
//			return nil
//		}
//		bucket = bucket.Bucket([]byte("Torrents"))
//		if bucket == nil {
//			return nil
//		}
//		v := bucket.Get([]byte(key))
//		if v == nil {
//			return nil
//		}
//		var t struct {
//			CreateDate time.Time `json:"CreateDate"`
//		}
//		if err := json.Unmarshal(v, &t); err == nil {
//			createDate = t.CreateDate
//		}
//		return nil
//	})
//
//	return createDate
//}
//
