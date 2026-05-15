package db

import (
	"NUMParser/db/models"
	"NUMParser/db/torrsearch"
	"NUMParser/utils"
	"github.com/agnivade/levenshtein"
	"sort"
	"strconv"
	"strings"
)

//func indexTorrs() {
//	if IsTorrsChange {
//		torrsearch.NewIndex(GetTorrs())
//	}
//}

func SearchTorr(query string) []*models.TorrentDetails {
	matchedIDs := torrsearch.Search(query)
	if len(matchedIDs) == 0 {
		return nil
	}
	torrs := GetTorrs()
	list := make([]*models.TorrentDetails, 0, len(matchedIDs))
	for _, id := range matchedIDs {
		list = append(list, torrs[id])
	}

	hash := utils.ClearStr(query)

	type sortKey struct {
		torr *models.TorrentDetails
		lev  int
	}
	keys := make([]sortKey, len(list))
	for i, t := range list {
		lhash := utils.ClearStr(strings.ToLower(t.Name+t.GetNames())) + strconv.Itoa(t.Year)
		keys[i] = sortKey{torr: t, lev: levenshtein.ComputeDistance(hash, lhash)}
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].lev == keys[j].lev {
			return keys[j].torr.CreateDate.Before(keys[i].torr.CreateDate)
		}
		return keys[i].lev < keys[j].lev
	})

	for i := range keys {
		list[i] = keys[i].torr
	}
	return list
}
