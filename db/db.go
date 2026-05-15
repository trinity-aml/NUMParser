package db

import (
	"NUMParser/db/db"
	"NUMParser/db/models"
	"NUMParser/db/rutor"
	"log"
)

func Init() {
	log.Println("Open db...")
	db.Init()

	log.Println("Read db...")
	rutor.Init()
}

func SaveAll() {
	rutor.SaveTorrs()
}

func GetTorrs() []*models.TorrentDetails {
	return rutor.GetTorrs()
}
