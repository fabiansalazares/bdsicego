// test decode .xer series into BDSICESerie structures.

package decode

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	dbLocalPath := "/home/alibey/econdata/bdsice/data/db/"

	var freqs = make(map[string]bool)

	files, err := ioutil.ReadDir(dbLocalPath)
	if err != nil {
		t.Error("DecodeFullDatabase: got error ", err.Error())
	}

	fileCounter := 0
	fileCounterTotal := len(files)

	for _, file := range files {

		fmt.Printf("Decoding test performed: %d\t\t%d\r", fileCounter, fileCounterTotal)
		fileCounter = fileCounter + 1

		if !strings.HasSuffix(file.Name(), ".xer") {
			continue
		}

		serie, err := Decode(dbLocalPath, file.Name())
		if err != nil {
			t.Error("Error at decode:", err.Error())
		}
		//t.Logf("Titulo: %s\n", serie.Title)
		if serie.SerieCode == "" {
			t.Errorf("%s: Expected a seriecode but got empty string.", serie.SerieCode)
		} else if serie.SerieCode == "tttt" {
			continue
		}

		if serie.Title == "" {
			t.Logf("%s: Expected a title but got empty string.", serie.SerieCode)
		}

		if serie.Units == "" {
			t.Errorf("%s: Expected a units but got empty string.", serie.SerieCode)
		}

		if serie.Source == "" {
			t.Errorf("%s: Expected a source but got empty string.", serie.SerieCode)
		}

		/* It's perfectly valid to not get any notes.
		if serie.Notes == nil {
			t.Errorf("%s: Expected notes but got empty string.", serie.SerieCode)
		}
		*/

		if serie.Frequency == 0 {
			t.Errorf("%s: Expected a frequency but got 0.", serie.SerieCode)
		} else {
			freqs[fmt.Sprintf("%d", serie.Frequency)] = true
		}

		/*
			if serie.Decimals == 0 {
				t.Error("Expected d.")
			}
		*/

		if serie.Start == nil {
			t.Errorf("%s: Expected a start but got empty string.", serie.SerieCode)
		}
		if serie.End == nil {
			t.Errorf("%s: Expected a end but got empty string.", serie.SerieCode)
		}
		if serie.NumberOfObservations == 0 {
			t.Errorf("%s: Expected a nob but got empty string.", serie.SerieCode)
		}

		/*
			if serie.Observations == econdataseries.Observations{} {
				t.Error("Expected a title but got empty string.")
			}
		*/

		if serie.Observations.Values == nil {
			t.Errorf("%s: observations values empty", serie.SerieCode)
		}

		if serie.Public == false {
			t.Errorf("%s: Expected public serie but got false.", serie.SerieCode)
		}

		if serie.Private == true {
			t.Errorf("%s: Expected not private but got true.", serie.SerieCode)
		}

		/*
			if serie.Active == 0 {
				t.Error("Expected a title but got empty string.")
			}

			if serie.Text == 0 {
				t.Error("Expected a title but got empty string.")
			}
		*/

	}
	for k, _ := range freqs {
		t.Logf("Freqs:\n%s\n", k)
	}
}
