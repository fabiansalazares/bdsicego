package database

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fabiansalazares/bdsicego/decode"
	"github.com/fabiansalazares/bdsicego/series"
	"io/ioutil"
)

func TestBuild(t *testing.T) {
	configuration, err := config.GetConfig()
	if err != nil {
		t.Errorf("TestBuild: could not load config file\n")

	}

	dbLocalPath := configuration.DatabaseLocalPath

	t.Logf("%s", dbLocalPath)

	filesInDirectory, err := ioutil.ReadDir(dbLocalPath)
	if err != nil {

		t.Errorf("TestBuild: error reading files to decode")
		//		return nil, fmt.Errorf("DecodeFullDatabase: %s", err.Error())
	}

	var filesToDecodePaths []string
	for _, fileToDecode := range filesInDirectory {
		if filepath.Ext(fileToDecode.Name()) == ".xer" {
			filesToDecodePaths = append(filesToDecodePaths, fileToDecode.Name())
		}
	}

	var seriesDecoded []*series.BDSICESerie
	var seriesDecodedCounter int
	var seriesToDecodeTotal = len(filesToDecodePaths)

	for _, fileToDecodePath := range filesToDecodePaths {
		// decode only .xer files
		if !strings.HasSuffix(fileToDecodePath, ".xer") {
			continue
		}

		fmt.Printf("Decoded: %d \tout of %d\r", seriesDecodedCounter, seriesToDecodeTotal)

		// log.Printf("Decoding %s\n", file.Name())

		serieToAdd, err := decode.Decode(dbLocalPath, path.Base(fileToDecodePath))
		if err != nil {
			t.Errorf("TestBuild: error adding serie to database")
		}

		seriesDecodedCounter += 1

		seriesDecoded = append(seriesDecoded, serieToAdd)
	}

	db, err := BuildDatabase(seriesDecoded)
	if err != nil {
		t.Errorf("TestBuild: Build returned an error %s", err.Error())
	}

	for code, serie := range db.Series {
		if serie == "" {
			log.Printf("Serie %s has an empty title", code)
			//t.("TestBuild: Found a nil serie in database")
		}
	}

}

func TestLoadDatabase(t *testing.T) {
	configuration, err := config.GetConfig()
	if err != nil {
		t.Errorf("TestBuild: could not load config file\n")

	}

	db, err := LoadDatabase(configuration)
	if err != nil {
		t.Errorf("TestLoadDatabase: error happened while loading: %s", err.Error())
	}

	if db == nil {
		t.Errorf("TestLoadDatabase: load returned a nil databasTestLoadDatabase: load returned a nil databasee")
	}
}

/*
func TestLoad(t *testing.T) {
	t.Logf("Testing Load()")

	configuration := config.GetConfig(config.GetDefaultConfigFilePath())

	db, err := LoadDatabase(configuration)
	if err != nil {
		t.Errorf("TestLoadDatabase: error happened while loading: %s", err.Error())
	}

	for _, serie := range db.GetCodes() {
		_, err = db.Load(serie)
		if err != nil {
			t.Fatalf("TestLoad(): could not load serie %s")
		}
	}
}
*/
