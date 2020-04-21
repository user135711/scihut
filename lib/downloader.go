package lib

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
)

// Most (if not all) of the trackers of the original torrents are no longer
// working, thus we maintain a list of trackers which are used by the
// community. The list of trackers does not require to be updated as often.
// See https://forum.mhut.org/viewtopic.php?f=17&t=8326
var TRACKERS = [][]string{
	{"udp://tracker.opentrackr.org:1337/announce"},
	{"udp://tracker.pirateparty.gr:6969/announce"},
	{"udp://zephir.monocul.us:6969/announce"},
	{"udp://tracker.tiny-vps.com:6969/announce"},
	{"udp://tracker.torrent.eu.org:451/announce"},
	{"udp://9.rarbg.to:2710/announce"},
	{"udp://tracker.coppersurfer.tk:6969/announce"},
	{"udp://tracker.cyberia.is:6969/announce"},
	{"udp://tracker.leechers-paradise.org:6969/announce"},
	{"http://tracker.gbitt.info/announce"},
	{"https://tracker.lelux.fi:443/announce"},
	{"udp://212.47.227.58:6969/announce"},
	{"udp://91.216.110.52:451/announce"},
}

// DownloadPaper downloads a research paper given:
//
//   - the infohash of the torrent it is in
//   - its DOI (Digital Object Identifier)
//   - its ID (assigned by Library Genesis)
func DownloadPaper(infoHash [20]byte, doi string, id uint64) ([]byte, error) {
	// Create a temporary directory for the torrent client
	tempdir, err := ioutil.TempDir("", "scihut-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempdir)

	// Create a new torrent client
	//   - Use the temporary storage we have just created
	//   - Listen on port 0 (i.e. randomized) so that consequent DownloadPapers
	//     will not fail due to port being (previously) in use
	config := torrent.NewDefaultClientConfig()
	config.DefaultStorage = storage.NewFileByInfoHash(tempdir)
	config.ListenPort = 0
	client, err := torrent.NewClient(config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Add the torrent by its infohash, add our trackers for faster discovery,
	// and wait until the metadata (info) is received. Finally, cancel all the
	// pieces as we will download selectively.
	torrent_, _ := client.AddTorrentInfoHash(infoHash)
	defer torrent_.Drop()
	torrent_.AddTrackers(TRACKERS)
	<-torrent_.GotInfo()
	log.Println("info: got torrent metadata")
	torrent_.CancelPieces(0, torrent_.NumPieces())
	log.Printf("info: connected to %d seeders\n", torrent_.Stats().ConnectedSeeders)

	// Find the ZIP file the PDF we are looking for might be in.
	zipFile := findZip(torrent_, id)
	if zipFile == nil {
		return nil, fmt.Errorf("could not find the zip file")
	}
	log.Printf("found zip file: %s\n", zipFile.Path())

	// Find the PDF file we are looking for.
	pdfFile, err := findPdf(zipFile, doi)
	if err != nil {
		return nil, err
	}
	log.Printf("found pdf file: %s\n", pdfFile.Name)

	// Read the PDF file.
	return readPdf(pdfFile)
}

func readPdf(pdfFile *zip.File) ([]byte, error) {
	pdfFileReader, err := pdfFile.Open()
	if err != nil {
		return nil, err
	}
	defer pdfFileReader.Close()
	return ioutil.ReadAll(pdfFileReader)
}

func findPdf(zipFile *torrent.File, doi string) (*zip.File, error) {
	zipFileReaderAt := NewReaderAt(zipFile.NewReader())
	zipReader, err := zip.NewReader(zipFileReaderAt, zipFile.Length())
	if err != nil {
		return nil, err
	}

	encodedDoi := encodeDoi(doi)
	for _, file := range zipReader.File {
		// Use ToUpper for case-insensitive comparison.
		// Remember that DOIs are case-insensitive!
		if strings.HasSuffix(strings.ToUpper(file.Name), strings.ToUpper(encodedDoi+".pdf")) {
			return file, nil
		}
	}

	return nil, fmt.Errorf("pdf not found")
}

func encodeDoi(doi string) string {
	tokens := strings.Split(doi, "/")
	return url.QueryEscape(tokens[0]) + "/" + url.QueryEscape(tokens[1])
}

func findZip(torrent_ *torrent.Torrent, id uint64) *torrent.File {
	re := regexp.MustCompile(`(?i).*libgen.scimag(\d{8})-(\d{8}).zip.*`)
	for _, file := range torrent_.Files() {
		matches := re.FindStringSubmatch(file.Path())
		if matches == nil {
			continue
		}
		lower, _ := strconv.ParseUint(matches[1], 10, 64)
		higher, _ := strconv.ParseUint(matches[2], 10, 64)

		// The range is inclusive.
		if lower <= id && id <= higher {
			return file
		}
	}
	return nil
}
