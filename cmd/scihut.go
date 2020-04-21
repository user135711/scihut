package main

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"scihut/lib"
)

func main() {
	if len(os.Args) != 3 {
		printUsage()
		os.Exit(1)
	}

	doi := os.Args[1]
	output := os.Args[2]
	if !validateDoi(doi) {
		log.Fatalln("error: invalid DOI")
	}
	if output == "-" {
		output = "/dev/stdout"
	}
	download(doi, output)
}

func printUsage() {
	usage := `scihut - decentralized access to scientific literature

Usage:
  scihut <doi> <output>`
	fmt.Println(usage)
}

func validateDoi(doi string) bool {
	return regexp.MustCompile(`(10[.][0-9]{4,}(?:[.][0-9]+)*/(?:\S)+)`).MatchString(doi)
}

func download(doi string, output string) {
	infoHash, id, err := lib.TranslateDoi(doi)
	if err != nil {
		log.Fatalf("error: could not translate the DOI: %s\n", err.Error())
	}
	log.Printf("info: paper has ID %d (group ID %03d) and can be found in torrent %x\n", id, id/100_000, infoHash)

	paperBlob, err := lib.DownloadPaper(infoHash, doi, id)
	if err != nil {
		log.Fatalf("error: could not download paper: %s\n", err.Error())
	}

	err = writePaper(paperBlob, output)
	if err != nil {
		log.Fatalf("error: could not write paper: %s\n", err.Error())
	}
}

func writePaper(paperBlob []byte, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	_, err = f.Write(paperBlob)
	return err
}
