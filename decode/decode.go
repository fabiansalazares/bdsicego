// package decode implements tools to decode .xer format into .json files
package decode

import (
	"bufio"
	"fmt"
	"time"

	"github.com/fabiansalazares/bdsicego/internal/utils"
	"github.com/fabiansalazares/bdsicego/series"

	"os"
	"path"
	"strconv"
	"strings"
)

// Generates a timestamp string from a .xer INI or FIN string, given a frequency
func xerTimeStringToTimestamp(xerString string, frequency int) (*time.Time, error) {
	parts := strings.Split(xerString, " ")

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("xerTimeStringToTimestamp(): %s\n", err.Error())
	}

	month := time.January // by default we assume values relate to the end of their corresponding period
	// Date() builds a Time object in such a manner that october 32 is nov 1
	// setting this to January we can set date values more easily taking advantage
	// of this behavior by Date()
	day := 1             // default unless frequency == 52 or == 365
	location := time.UTC // default timezone

	switch frequency {
	case 1:
	case 4:
		monthNumber, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("xerTimeStringToTimestamp(): %s\n", err.Error())
		}
		month = time.Month(monthNumber * 3)

	case 12:
		monthNumber, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("xerTimeStringToTimestamp(): %s\n", err.Error())
		}
		month = time.Month(monthNumber)
	case 52:
		// so much work for just two series 676020 and 620120
		monthNumber, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("xerTimeStringToTimestamp(): %s\n", err.Error())
		}
		month = time.Month(monthNumber)

		day, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("xerTimeStringToTimestamp(): %s\n", err.Error())
		}

		//    week is third character      day in the week is second
	case 365:
		monthNumber, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("xerTimeStringToTimestamp: %s\n", err.Error())
		}

		month = time.Month(monthNumber)

		day, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("xerTimeStringToTimestamp: %s\n", err.Error())
		}

	}
	timeObject := time.Date(year, month, day, 0, 0, 0, 0, location)

	return &timeObject, nil
}

// decodes the .xer file in dbLocalPath and returns a pointer to a BDSICESerie struct
// contaning all the fields in the .xer file
func Decode(dbLocalPath string, XerFile string) (*series.BDSICESerie, error) {

	s := series.BDSICESerie{}
	s.Public = true
	s.Private = false

	fileToOpen := path.Join(dbLocalPath, XerFile)
	//fmt.Printf("About to open %s\n", fileToOpen)

	file, err := os.Open(fileToOpen)
	if err != nil {
		return nil, fmt.Errorf("Decode(): %s", err.Error())
	}
	defer file.Close()

	//defer file.Close()
	s.SerieCode = strings.Split(path.Base(XerFile), ".")[0]
	//fmt.Printf("SerieCode: %s\n", s.SerieCode)
	reader := bufio.NewReader(file)

	var line string
	var lineCount int

	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}

		if strings.HasPrefix(line, "COD:") {
			lineCount++

		} else if strings.HasPrefix(line, "TIT:") { // extract title of the serie
			split := strings.Split(line, "TIT: ")

			trimTitle := strings.TrimSuffix(split[1], "\r\n")

			trimTitle = utils.ReplaceSpanishCharacters(trimTitle)
			s.Title = trimTitle
			lineCount++
		} else if strings.HasPrefix(line, "UNI: ") { // extract units
			s.Units = strings.TrimSuffix(strings.Split(line, "UNI:")[1], "\r\n")
			lineCount++

		} else if strings.HasPrefix(line, "FUE: ") { // extract the source of the serie
			s.Source = strings.TrimSuffix(strings.Split(line, "FUE:")[1], "\r\n")
			lineCount++

		} else if strings.HasPrefix(line, "NOT: ") {
			var noteLine string

			for {
				noteLine, err = reader.ReadString('\n')
				if err != nil {
					break
				}
				lineCount++

				if strings.HasPrefix(noteLine, "@") {
					break
				} else {
					s.Notes = append(s.Notes, noteLine)
				}
			}
			lineCount++

		} else if strings.HasPrefix(line, "DEC:") { // extract number of decimals in the serie
			decParts := strings.Split(line, "DEC: ")

			decimals, err := strconv.Atoi(strings.TrimSuffix(decParts[1], "\r\n"))
			if err != nil {
				return nil, fmt.Errorf("Decode(): %s", err.Error())
			}

			s.Decimals = decimals

			lineCount++

		} else if strings.HasPrefix(line, "FRE:") { // extract frequency

			freqParts := strings.Split(strings.TrimSuffix(strings.Trim(strings.Split(line, "FRE: ")[1], " "), "\r\n"), " ")

			if len(freqParts) < 2 {
				f, err := strconv.Atoi(freqParts[0])
				if err != nil {
					return nil, fmt.Errorf("Decode(): %s", err.Error())
				}
				s.Frequency = f
			} else {
				// All frequencies are then FES: 3
				//fmt.Printf("%q\n", freqParts)
				f, err := strconv.Atoi(freqParts[0])
				if err != nil {
					return nil, fmt.Errorf("Decode(): %s", err.Error())
				}
				s.Frequency = f

			}

			/*
				freqParts = strings.TrimSuffix(freqParts[0], " ")

				f, err := strconv.Atoi(strings.Trim(freqParts[1], "\r\n", " "))
				if err != nil {
					return nil, fmt.Errorf("Decode: could not convert frequency %s", line)
				}

				s.Frequency = f */
			lineCount++

		} else if strings.HasPrefix(line, "INI:") { // extract starting period
			//s.Start = strings.TrimSpace(strings.TrimSuffix(strings.Split(line, "INI:")[1], "\r\n"))

			s.Start, err = xerTimeStringToTimestamp(strings.TrimSpace(strings.TrimSuffix(strings.Split(line, "INI:")[1], "\r\n")), s.Frequency)
			if err != nil {
				return nil, fmt.Errorf("Decode(): %s", err.Error())
			}
			lineCount++

		} else if strings.HasPrefix(line, "FIN:") { // extract ending period
			s.End, err = xerTimeStringToTimestamp(strings.TrimSpace(strings.TrimSuffix(strings.Split(line, "FIN:")[1], "\r\n")), s.Frequency)
			if err != nil {
				return nil, fmt.Errorf("Decode(): %s", err.Error())
			}
			lineCount++
		} else if strings.HasPrefix(line, "NOB:") { // extract number of observations
			var nobLine string

			nobParts := strings.Split(line, "NOB: ")

			nob, err := strconv.Atoi(strings.TrimSuffix(nobParts[1], "\r\n"))
			if err != nil {
				return nil, fmt.Errorf("Decode: NOB: decimal field %q could not be converted to integer.", nobParts[1])
			}

			s.NumberOfObservations = nob

			for {
				nobLine, err = reader.ReadString('\n')
				if err != nil {
					break
				}
				lineCount++

				if strings.HasPrefix(nobLine, "PUB:") || strings.HasPrefix(nobLine, "#") {
					// true by default, nothing to do
					lineCount++
					break
				}

				for _, obs := range strings.Split(nobLine, " ") {
					obs = strings.TrimSpace(obs)

					if obs == "" {
						continue
					} else if obs == "OM" || obs == "ND" {
						// JSON marshal does not support math.NaN values
						// The workaround here is to append a very unlikely value such  as
						// -9999999.9999999 and toggle a field called ContainsNan
						// When BDSICESerie.Data() gets implemented, it will have to check for
						// ContainsNaN and substitute all -999999999.99999 values for a math.NaN
						s.Observations.Values = append(s.Observations.Values, -99999999.999999)
						s.ContainsNan = true
					} else {
						value, err := strconv.ParseFloat(obs, 64)
						if err != nil {
							return nil, fmt.Errorf("Decode(): obs %s could not be converted to float64", obs)
						}
						s.Observations.Values = append(s.Observations.Values, value)
					}
				}

			}

			if strings.HasPrefix(nobLine, "#") {
				break
			}
		} else if strings.HasPrefix(line, "PRI:") { // extract value signalling a private serie (always false in practice)
			// false by default, nothing to do
			lineCount++

		} else if strings.HasPrefix(line, "DET:") {
			detParts := strings.Split(line, ": ")
			det, err := strconv.Atoi(strings.TrimSuffix(detParts[1], "\r\n"))
			if err != nil {
				return nil, fmt.Errorf("Decode(): DET field: unable to convert %s", detParts[1])
			}

			if det == 0 {
				s.Active = false
			} else if det == 1 {
				s.Active = true
			} else {
				return nil, fmt.Errorf("Decode(): DET field is %s, must be 0 or 1", detParts[1])
			}
			lineCount++

		} else if strings.HasPrefix(line, "TEX:") { // extract text field
			var textLine string

			for {
				textLine, err = reader.ReadString('\n')
				if err != nil {
					break
				}
				lineCount++

				if strings.HasPrefix(textLine, "#") {
					break
				} else {
					s.Notes = append(s.Text, textLine)
				}
			}
			//} else if strings.HasPrefix(line, "#") {
			//	break
		} else if strings.HasPrefix(line, "#") { // # signals the end of a .xer file
			//fmt.Printf("About to break?")
			break
		} else {
			lineCount++
			return nil, fmt.Errorf("Decode(): unrecognized line:\n%s\nat .xer file %d", line, lineCount)

		}
	}

	// Generate the timestamps according to .Frequency, .Start and .End
	// We got possible frequencies 1,4,12,52,365

	var dates = make([]time.Time, s.NumberOfObservations)
	dates[0] = (*s.Start)

	switch s.Frequency {
	case 1:
		// add exactly one year to the previous date and add it to dates[]
		for i := 1; i < s.NumberOfObservations; i++ {
			dates[i] = dates[i-1].AddDate(1, 0, 0)
		}

	case 4:
		for i := 1; i < s.NumberOfObservations; i++ {
			dates[i] = dates[i-1].AddDate(0, 3, 0)
		}
	case 12:

		for i := 1; i < s.NumberOfObservations; i++ {
			dates[i] = dates[i-1].AddDate(0, 1, 0)
		}
	case 52:

		for i := 1; i < s.NumberOfObservations; i++ {
			dates[i] = dates[i-1].AddDate(0, 0, 7)
		}
	case 365:

		for i := 1; i < s.NumberOfObservations; i++ {
			dates[i] = dates[i-1].AddDate(0, 0, 1)
		}
	}

	s.Observations.Dates = dates

	return &s, nil

}
