package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const DATABASE_PATH = "assets/database.sqlite3"
const TSV_PATH = "assets/scimag_files.tsv"

func main() {
	csvFile, err := os.Open(TSV_PATH)
	if err != nil {
		log.Fatalln(err)
	}
	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = '\t'
	csvReader.LazyQuotes = true

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_journal=MEMORY&_locking=EXCLUSIVE&_txlock=exclusive", DATABASE_PATH))
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}

	_, err = db.Exec(`
		CREATE TABLE scimag_files (
			id INTEGER PRIMARY KEY,
			md5 TEXT NOT NULL COLLATE NOCASE,
			doi TEXT NOT NULL COLLATE NOCASE,
			filesize INTEGER NOT NULL
		);
	`)
	if err != nil {
		log.Fatalln(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO scimag_files (id, md5, doi, filesize) VALUES (?, ?, ?, ?);`)
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	log.Println("importing...")
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

		id := record[0]
		md5 := record[1]
		doi := record[2]
		filesize, err := strconv.ParseUint(record[3], 10, 64)
		if err != nil {
			log.Fatalln(err)
		}

		_, err = stmt.Exec(id, md5, doi, filesize)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if err = tx.Commit(); err != nil {
		log.Fatalln(err)
	}

	log.Println("creating index on doi...")
	_, err = db.Exec(`CREATE INDEX scimag_files__doi ON scimag_files (doi COLLATE NOCASE);`)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("all OK!")
}
