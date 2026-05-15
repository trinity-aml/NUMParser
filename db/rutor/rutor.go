package rutor

import (
	"NUMParser/config"
	"NUMParser/db/db"
	"NUMParser/db/models"
	"NUMParser/db/torrsearch"
	"NUMParser/utils"
	"compress/flate"
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	torrs         []*models.TorrentDetails
	hashIndex     map[string]int
	IsTorrsChange bool
	muTorrs       sync.Mutex
)

func Init() {
	db.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Rutor"))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte("Torrents"))
		if bucket == nil {
			return nil
		}
		stats := bucket.Stats()
		if stats.KeyN > 0 {
			torrs = make([]*models.TorrentDetails, 0, stats.KeyN)
		}
		err := bucket.ForEach(func(_, v []byte) error {
			var torr *models.TorrentDetails
			err := json.Unmarshal(v, &torr)
			if err == nil {
				torrs = append(torrs, torr)
			}
			return err
		})
		if err != nil {
			log.Println("Error read rutor from db:", err)
		}
		return nil
	})

	rebuildHashIndex()
	torrsearch.NewIndex(torrs)
}

func rebuildHashIndex() {
	hashIndex = make(map[string]int, len(torrs))
	for i, t := range torrs {
		if t.Hash != "" {
			hashIndex[t.Hash] = i
		}
	}
}

func RemoveAll() {
	torrs = nil
	hashIndex = nil
	db.DB.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("Rutor"))
		return nil
	})
}

func GetTorrs() []*models.TorrentDetails {
	return torrs
}

func SetTorrs(list []*models.TorrentDetails) {
	torrs = list
	rebuildHashIndex()
	IsTorrsChange = true
}

func AddTorr(t *models.TorrentDetails) {
	muTorrs.Lock()
	defer muTorrs.Unlock()

	if hashIndex == nil {
		hashIndex = make(map[string]int, len(torrs)+1)
	}

	if t.Hash != "" {
		if i, ok := hashIndex[t.Hash]; ok {
			t.IMDBID = torrs[i].IMDBID
			torrs[i] = t
			return
		}
	}

	IsTorrsChange = true
	if t.Hash != "" {
		hashIndex[t.Hash] = len(torrs)
	}
	torrs = append(torrs, t)
}

func SaveTorrs() {
	removeOldTorr()
	if !IsTorrsChange || len(torrs) == 0 {
		return
	}
	muTorrs.Lock()
	defer muTorrs.Unlock()
	log.Println("Save torrents")

	err := db.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Rutor"))
		if err != nil {
			return err
		}
		//Recreate torrents
		bucket.DeleteBucket([]byte("Torrents"))
		bucket, err = bucket.CreateBucket([]byte("Torrents"))
		if err != nil {
			return err
		}

		for _, torr := range torrs {
			buf, err := json.Marshal(torr)
			if err != nil {
				return err
			}
			err = bucket.Put([]byte(strings.ToLower(torr.Hash)), buf)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalln("Error write to db rutor:", err)
	}
	saveRutorLS()
	torrsearch.NewIndex(torrs)
	IsTorrsChange = false
}

func removeOldTorr() {
	muTorrs.Lock()
	defer muTorrs.Unlock()

	list := make([]*models.TorrentDetails, 0, len(torrs))
	inResult := make(map[string]struct{}, len(torrs))

	for _, t := range torrs {
		if _, ok := inResult[t.Link]; !ok {
			inResult[t.Link] = struct{}{}
			list = append(list, t)
		}
	}

	torrs = list
	rebuildHashIndex()
}

func saveRutorLS() {
	log.Println("Save torrents rutor.ls")
	dir := filepath.Dir(os.Args[0])
	ff, err := os.Create(filepath.Join(dir, "rutor.ls"))
	if err != nil {
		log.Println("Error save torrs:", err)
		return
	}
	defer ff.Close()

	w, err := flate.NewWriter(ff, flate.BestCompression)
	if err != nil {
		log.Println("Error save torrs:", err)
		return
	}
	defer w.Close()

	rutorHost := config.RutorHost()
	out := make([]*models.TorrentDetails, len(torrs))
	for i, t := range torrs {
		cp := *t
		cp.Link = config.JoinRutorLink(rutorHost, cp.Link)
		out[i] = &cp
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(out)
	if err != nil {
		log.Println("Error save torrs:", err)
		return
	}

	src := dir + "/rutor.ls"
	dest := dir + "/public/releases/rutor.ls"
	err = utils.CopyFile(src, dest)
	if err != nil {
		log.Println("Error copy rutor.ls:", err)
		return
	}
}
