// Testing file for bdsicego/series
// We only test function Load() because all the other functions are currently too simple to fail

package series

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fabiansalazares/bdsicego/internal/config"
)

func TestLoad(t *testing.T) {
	configuration, err := config.GetConfig()
	if err != nil {
		t.Errorf("TestLoad: %s", err.Error())
	}

	filesToAppend, err := ioutil.ReadDir(configuration.DatabaseLocalPath)
	if err != nil {
		t.Fatalf("ioutil.ReadDir(configuration.DatabaseLocalPath): %s", err.Error())
	}

	var seriesToLoad []string
	for _, fileToAppend := range filesToAppend {
		if filepath.Ext(fileToAppend.Name()) == ".json" {
			serieToAppend := strings.TrimSuffix(fileToAppend.Name(), ".json")
			seriesToLoad = append(seriesToLoad, serieToAppend)
			t.Logf("Appending %s\n", serieToAppend)
		}
	}

	for _, serieToLoad := range seriesToLoad {
		serie, err := Load(configuration, serieToLoad)
		if err != nil {
			t.Fatalf("An error happened while loading serie %s: %s", serie, err.Error())
		}
	}

}
