// Package database implements tools and structures to search and access the series in the BDSICE.
package database

import (
	"unicode"

	"github.com/fabiansalazares/bdsicego/internal/config"
	"github.com/fabiansalazares/bdsicego/series"

	// "econdata/internal/bdsice/utils"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	//	"fmt"
	"time"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// checks for unicode separator characters. Used by searchCommand to normalise search terms.
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

type BDSICEDatabase struct {
	Series       map[string]string `json:"Series"` // code: title
	LastUpdate   time.Time         `json:"LastUpdate"`
	DatabasePath string            `json:"Path"`
	Codes        []string          `json:"Codes"`
}

// returns the series that contain all of the terms either in the title or in the serie code
// Match() does not normalize the search terms and the calling function should be responsible of performing any normalisation of diacritics and others that might me deemed necessary. Normalisations at this levels would reduce performance.
func (db *BDSICEDatabase) Match(terms ...string) map[string]string { // []*BDSICEDatabaseSerie {

	var results map[string]string
	results = make(map[string]string)

	if len(terms) == 0 {
		return nil
	} else if len(terms) == 1 {
		for code, title := range db.Series {
			if strings.Contains(code, terms[0]) || strings.Contains(title, strings.ToUpper(terms[0])) {
				results[code] = title
			}
		}
		return results
	} else {
		resultsOtherTerms := db.Match(terms[1:]...)

		for code, title := range resultsOtherTerms {
			if strings.Contains(code, terms[0]) || strings.Contains(title, strings.ToUpper(terms[0])) {
				results[code] = title
			}
		}
	}

	return results
}

// returns the series that contain all the terms either in the title or the serie code, and excludes all the series that contain the terms prefixed by "-" either in the title or the serie code.
func (db *BDSICEDatabase) Search(terms ...string) (map[string]string, error) {

	var (
		matchTerms   []string // terms that we will match the database against
		excludeTerms []string // terms that are prefixed by "-": series that contain them in title or code will be excluded from results
		cleanTerms   []string
	)

	transformChain := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)

	for _, term := range terms {
		termNormalized, _, err := transform.String(transformChain, term)
		if err != nil {
			return nil, fmt.Errorf("database.Search(): %s", err.Error())
		}

		//		termNormalizedUpper := strings.ToUpper(termNormalized)

		//		cleanTerms = append(cleanTerms, termNormalizedUpper)

		cleanTerms = append(cleanTerms, termNormalized)
	}

	//fmt.Printf("cleanTerms: %s\n", cleanTerms)

	for _, term := range cleanTerms {
		if strings.HasPrefix(term, "-") {
			excludeTerms = append(excludeTerms, term[1:])
		} else {
			matchTerms = append(matchTerms, term)
		}
	}

	matchResults := db.Match(matchTerms...)

	if len(excludeTerms) > 0 {

		fmt.Printf("Terms to exclude: %s\n", excludeTerms)
		for _, termToExclude := range excludeTerms {
			for code, title := range matchResults {
				if strings.Contains(code, termToExclude) || strings.Contains(title, strings.ToUpper(termToExclude)) {
					delete(matchResults, code)
				}
			}
		}

	}
	return matchResults, nil

}

// Adds a BDSICEDatabaseSerie object to a BDSICEDatabase from a BDSICESerie
// Basically, it takes SerieCode and Title fields of BDSICESerie and appends them
// to the array of BDSICEDatabaseSerie objects
func (db *BDSICEDatabase) AddSerie(serie *series.BDSICESerie) error {

	//	fmt.Printf("Adding SerieCode: %s\nAdding title: %s\n", serie.SerieCode, serie.Title)

	/*	db.Series = append(db.Series, &BDSICEDatabaseSerie{
		SerieCode: serie.SerieCode,
		Title:     serie.Title})*/

	db.Series[serie.SerieCode] = serie.Title

	return nil

}

/*
// Load the full series given as codes from the database object
// This is one of methods that implements the interface econseries.Econserie
func (d *BDSICEDatabase) Load(codes ...string) ([]*series.EconSerie, error) {
	return nil
}
*/

// returns a string slice containing a list of all the codes in the database
func (db *BDSICEDatabase) GetCodes() []string {
	return db.Codes
}

// builds a db.json file from a BDSICEDatabase object from an array of pointers of series.BDSICESerie objects
// Intented to be called from download.DownloadFullDatabase() or download.Update() at some point after
// having downloaded or updated the database of .json files
func BuildDatabase(seriesToBuild []*series.BDSICESerie) (*BDSICEDatabase, error) {

	var db BDSICEDatabase

	db.Series = make(map[string]string)
	db.Codes = make([]string, 0, len(seriesToBuild))

	// fmt.Println("Building database...")
	for _, serie := range seriesToBuild {
		db.Codes = append(db.Codes, serie.SerieCode)
		err := db.AddSerie(serie)

		if err != nil {
			return nil, fmt.Errorf("database.BuildDatabase(): %s", err.Error())
		}
	}

	db.LastUpdate = time.Now()

	//	fmt.Printf("Printing a random db element: \nserieCode: %s\ntitle: %s\n", db.series[1452].serieCode, db.series[1452].title)

	// MARSHALL DB TO JSON AND WRITE TO FILE BDSICEdb.json

	return &db, nil
}

// loads a database from database path specified in configuration and returns
// a reference to a database object
func LoadDatabase(configuration *config.BDSICEConfig) (*BDSICEDatabase, error) {
	var db BDSICEDatabase
	db.Series = make(map[string]string)
	dbJsonFilePath := path.Join(configuration.DatabaseLocalPath, "db.json")

	// fmt.Printf("Loading db from %s\n", dbJsonFilePath)

	dbFileReader, err := ioutil.ReadFile(dbJsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("database.LoadDatabase(): %s", err.Error())
	}

	err = json.Unmarshal([]byte(dbFileReader), &db)
	if err != nil {
		return nil, fmt.Errorf("database.LoadDatabase(): %s", err.Error())
	}

	return &db, nil
}
