// package series defines data structures and functions to manipulate and retrieve BDSICE series
package series

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/fabiansalazares/bdsicego/internal/config"
)

type EconSerie interface {
	GetData() *Observations
	String() string
	GetCode() string
	GetTitle() string
	GetUnit() string
	Average() float64
	Min() (float64, time.Time) // must return the minimum observation in the serie and its time
	Max() (float64, time.Time) // must return the maximum value in the serie and its time
}

type Observations struct {
	Dates  []time.Time `json:"Dates"`
	Values []float64   `json:"Values"`
}

// TODO:
// Ver ffjson https://github.com/pquerna/ffjson

type BDSICESerie struct {
	SerieCode            string       `json:"SerieCode"`
	Title                string       `json:"Title"`
	Units                string       `json:"Units"`
	Source               string       `json:"Source"`
	Notes                []string     `json:"Notes"`
	Decimals             int          `json:"Decimals"`
	Frequency            int          `json:"Frequency"`
	Start                *time.Time   `json:"Start"`
	End                  *time.Time   `json:"End"`
	NumberOfObservations int          `json:"NumberOfObservations"`
	Observations         Observations `json:"Observations"`
	Public               bool         `json:"Public"`
	Private              bool         `json:"Private"`
	Active               bool         `json:"Active"`
	Text                 []string     `json:"Text"`
	ContainsNan          bool         `json:"ContainsNaN"`
}

// returns a string representation of the serie
func (s BDSICESerie) String() string {
	return fmt.Sprintf("Serie: BDSICE -- %s -- %s", s.SerieCode, s.Title)
}

// JSON marshal does not support math.NaN values
// The workaround here is to append a very unlikely value such  as
// -9999999.9999999 and toggle a field called ContainsNan
// When BDSICESerie.Data() gets implemented, it will have to check for
// ContainsNaN and substitute all -999999999.99999 values for a math.NaN
func (s BDSICESerie) GetData() *Observations { return &s.Observations }

// returns a string containing the code of the serie
func (s BDSICESerie) GetCode() string { return s.SerieCode }

// returns a string containing the title of the serie
func (s BDSICESerie) GetTitle() string { return s.Title }

func (s BDSICESerie) GetUnit() string { return s.Units }

// returns a float64 number containing the average value of the full serie
func (s BDSICESerie) Average() float64 {
	obs := s.Observations

	var sum float64

	for _, ob := range obs.Values {
		sum += ob
	}

	return (sum / float64(len(obs.Values)))

}

// returns a float64 number containing the minimum value in the serie and its associated value
func (s BDSICESerie) Min() (float64, time.Time) {
	var minValue float64
	var minTime time.Time

	data := s.Observations

	minValue = data.Values[0]
	minTime = data.Dates[0]

	for i := 1; i < len(data.Values); i++ {
		if data.Values[i] < minValue {
			minValue = data.Values[i]
			minTime = data.Dates[i]
		}
	}

	return minValue, minTime
}

// returns a float64 number contaning the maximum value in the serie and its associated date
func (s BDSICESerie) Max() (float64, time.Time) {
	var maxValue float64
	var maxTime time.Time

	obs := s.Observations

	maxValue = obs.Values[0]
	maxTime = obs.Dates[0]

	for i := 1; i < len(obs.Values); i++ {
		if obs.Values[i] > maxValue {
			maxValue = obs.Values[i]
			maxTime = obs.Dates[i]
		}
	}

	return maxValue, maxTime
}

// loads the serie corresponding to serieCode in dbLocalPath and returns a pointer to the BDSICESerie
// object.
func Load(configuration *config.BDSICEConfig, serieCode string) (*BDSICESerie, error) {

	var serie BDSICESerie

	serieJsonFilePath := path.Join(configuration.DatabaseLocalPath, fmt.Sprintf("%s.json", serieCode))

	if configuration.Debug {
		fmt.Printf("Loading BDSICESerie from %s\n", serieJsonFilePath)
	}

	serieFileReader, err := ioutil.ReadFile(serieJsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("series: Load(): reading file %s: %s", serieJsonFilePath, err.Error())
	}

	err = json.Unmarshal([]byte(serieFileReader), &serie)
	if err != nil {
		return nil, fmt.Errorf("database: Load(): unmarshaling JSON: %s", err.Error())
	}

	return &serie, nil

}
