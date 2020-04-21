.PHONY: update_database update_torrents

bin/scihut: cmd/* lib/*
	go build -o ./bin/scihut ./cmd/scihut.go

bin/update_database: util/update_database/main.go
	go build -o ./bin/update_database ./util/update_database/main.go

assets/scimag_files.tsv.bz2:
	@echo "WARNING: the download is slow and unreliable, patience and insistence!"
	wget --continue "http://gen.lib.rus.ec/dbdumps/scimag-data/scimag_files.tsv.bz2" -O assets/scimag_files.tsv.bz2

assets/scimag_files.tsv: assets/scimag_files.tsv.bz2
	bzip2 --decompress --keep --stdout assets/scimag_files.tsv.bz2 > assets/scimag_files.tsv

assets/database.sqlite3: update_database

assets/torrents.json: update_torrents


# PHONY targets
# =============
update_database: assets/scimag_files.tsv util/update_database/main
	./bin/update_database

update_torrents:
	python3 ./util/update_torrents/