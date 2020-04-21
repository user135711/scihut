package lib

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const DatabasePath = "assets/database.sqlite3"
const TorrentsPath = "assets/torrents.json"

// TranslateDoi "translates" a DOI (Digital Object Identifier) to the following
//
//   - InfoHash (of the torrent the paper can be found in)
//   - ID (of the paper as assigned by Library Genesis)
//
// and also returns an error (if any). Its return values are used by
// DownloadPaper.
func TranslateDoi(doi string) ([20]byte, uint64, error) {
	// For testing without a database
	if strings.EqualFold(doi, "10.1002/(sici)1097-4628(19960425)60:4<531::aid-app6>3.0.co;2-p") {
		ih, _ := hex.DecodeString("99663320037082381eab3876cd981c6941c2a656")
		var infoHash [20]byte
		copy(infoHash[:], ih)
		return infoHash, 54296, nil
	}

	id, err := doiToId(doi)
	if err != nil {
		return [20]byte{}, 0, err
	}

	infoHash, err := idToInfoHash(id)
	if err != nil {
		return [20]byte{}, 0, err
	}

	return infoHash, id, nil
}

func doiToId(doi string) (uint64, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?mode=ro", DatabasePath))
	if err != nil {
		return 0, err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return 0, err
	}

	rows, err := db.Query(`SELECT id FROM scimag_files WHERE doi = ? COLLATE NOCASE`, doi)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, fmt.Errorf("no ID found for the DOI")
	}

	var id uint64
	if err = rows.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func idToInfoHash(id uint64) ([20]byte, error) {
	f, err := os.Open(TorrentsPath)
	if err != nil {
		return [20]byte{}, err
	}
	defer f.Close()

	var torrents map[string]string
	err = json.NewDecoder(f).Decode(&torrents)
	if err != nil {
		return [20]byte{}, err
	}

	infoHashHex, ok := torrents[fmt.Sprintf("%d", id/100_000)]
	if !ok {
		return [20]byte{}, fmt.Errorf("infohash for the ID %d (ID-Group %d) not found", id, id/100_000)
	}

	decodedHex, err := hex.DecodeString(infoHashHex)
	if err != nil {
		return [20]byte{}, fmt.Errorf("could not decode infohash hex (%s); torrent list corrupt", infoHashHex)
	}
	var ih [20]byte
	copy(ih[:], decodedHex)
	return ih, nil
}
