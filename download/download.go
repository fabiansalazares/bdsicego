// functions to download the BDSICE databases, fetch updates and get the latest coyuntura bulletins
package download

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"gibhub.com/fabiansalazares/bdsicego/series"
	"github.com/fabiansalazares/bdsicego/database"

	"net/http/httputil"

	"net/url"
	"os"
	"path"
	"path/filepath"

	"gibhub.com/fabiansalazares/bdsicego/decode"

	//	"bytes"
	//	"encoding/json"

	"strconv"
	"strings"

	"golang.org/x/net/html"

	"gibhub.com/fabiansalazares/bdsicego/internal/config"
)

// copies from response to destination providing a visual lead on how much remains to be downloaded
func downloadWithCounter(configuration *config.BDSICEConfig, response *http.Response) (string, int64, error) {
	var byteCounter int64
	var zipFileName string
	var zipFileLength int
	var zipFileLengthString string
	var zipFilePath string

	// Check that there is actually a file waiting to be downloaded
	// We do this by checking that there is a Content-Disposition header, and that the
	// header includes a 'filename' variable.
	if _, ok := response.Header["Content-Disposition"]; !ok {
		log.Fatalln("BDSICE's response does not include a zip file to download.")
	}

	if headerFileNameParts := strings.Split(response.Header["Content-Disposition"][0], "filename="); len(headerFileNameParts) > 1 {
		zipFileName = headerFileNameParts[1]
	} else {
		log.Fatalln("No filename in Content-Disposition.")
	}

	// Join local database path and zip file name to get full path of database zip file
	zipFilePath = path.Join(configuration.DatabaseLocalPath, zipFileName)

	// Extract content size by reading Content-Length header
	zipFileLength, err := strconv.Atoi(response.Header["Content-Length"][0])
	if err != nil {
		log.Printf("Content-Length could not be read. Assuming length unknown")
		// if we could not determine, we set the length string to "?K"
		zipFileLengthString = "?K"
	} else {
		zipFileLengthString = fmt.Sprintf("%d", (zipFileLength / 1024))
	}

	// Create an empty file at zipFilePath and return a file handler
	out, err := os.Create(zipFilePath)
	if err != nil {
		log.Fatal(err)
	}

	defer out.Close()

	// Download from response body into the zip file that we just created
	fmt.Printf("Downloading...")
	for {
		bytesCopied, err := io.CopyN(out, response.Body, 1024*100)
		if err != nil {
			fmt.Printf("\rCopied: %sK\tTotal: %sK\n", zipFileLengthString, zipFileLengthString)
			fmt.Printf("Download complete.\n")
			fmt.Printf("\n")
			break
		} else {
			byteCounter = byteCounter + bytesCopied
			fmt.Printf("\rCopied %dK\tTotal: %sK\r", (byteCounter / 1024), zipFileLengthString)
		}
	}

	return zipFilePath, byteCounter, err
}

// check if an update has been applied already. This is somewhat of a placeholder
// function. Currently it just checks if a folder named daymonthyear exists
// and returns true if it exists. It should check against somekind of database
// that tracked the integrity of the database and the updates that have been
// applied.
func alreadyDownloadedUpdate(updatePath string, day int, month int, year int) bool {

	pathToCheck := path.Join(updatePath, fmt.Sprintf("UltActualiz_%04d%02d%02d", year, month, day))
	fmt.Printf("Checking path: %s\n", pathToCheck)

	if _, err := os.Stat(pathToCheck); !os.IsNotExist(err) {
		// it does exists, thus update has been downloaded already
		return true
	} else {
		fmt.Printf("There is an update that has not been downloaded yet.\n")
		return false
	}
}

// Should check if full db exists and ask user for confirmation
// PLACEHOLDER for now
func alreadyDownloadedFullDatabase(configuration *config.BDSICEConfig, databasePath string) bool {

	if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
		fmt.Printf("Checking %s\n", filepath.Join(configuration.DatabaseLocalPath, "db.json"))

		if _, err := os.Stat(filepath.Join(configuration.DatabaseLocalPath, "db.json")); !os.IsNotExist(err) {
			return true
		}
	}

	return false
}

// returns DOM elements by id
func getElementById(id string, n *html.Node) (element *html.Node, ok bool) {
	for _, a := range n.Attr {
		if a.Key == "id" && a.Val == id {
			return n, true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if element, ok = getElementById(id, c); ok {
			return
		}
	}
	return
}

// unzips origin file to destination path, creating a folder if necessary
func unzipFile(origin string, destination string, directExtract bool) ([]string, error) {
	var extractedFiles []string
	var extractDirPath string

	// a directory with the base of the origin minus the extension will be
	// created in destination path if it does not exist already.
	// files in the zip file will be extracted to that directory
	if directExtract {
		extractDirPath = destination
	} else {
		extractDirPath = path.Join(destination, strings.Split(path.Base(origin), ".")[0])
	}
	// check directory exists already. If not, create.
	if _, err := os.Stat(extractDirPath); os.IsNotExist(err) {
		os.Mkdir(extractDirPath, 0755) // rwxr-xr-x
	}

	r, err := zip.OpenReader(origin)
	if err != nil {
		return extractedFiles, err
	}
	defer r.Close()

	for _, f := range r.File {
		fileToExtractPath := filepath.Join(extractDirPath, f.Name)

		// fmt.Printf("Checking ZipSlip for: \nfileToExtractPath: %s\n%s\n", fileToExtractPath, filepath.Clean(extractDirPath)+string(os.PathSeparator))
		// check for ZipSlip vulnerability.
		if !strings.HasPrefix(fileToExtractPath, filepath.Clean(extractDirPath)+string(os.PathSeparator)) {
			log.Fatalf("%s: illegal file path. Possible ZipSlip attack.\n", fileToExtractPath)
		}

		extractedFiles = append(extractedFiles, fileToExtractPath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fileToExtractPath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fileToExtractPath), os.ModePerm); err != nil {
			return extractedFiles, err
		}

		outFile, err := os.OpenFile(fileToExtractPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return extractedFiles, err
		}

		zippedFileHandler, err := f.Open()
		if err != nil {
			return extractedFiles, err
		}

		_, err = io.Copy(outFile, zippedFileHandler)
		outFile.Close()
		zippedFileHandler.Close()

		if err != nil {
			return extractedFiles, err
		}
	}

	return extractedFiles, nil
}

func DecodePartialDatabase(configuration *config.BDSICEConfig, filesToDecode []string) ([]*series.BDSICESerie, error) {
	var seriesDecoded []*series.BDSICESerie
	var counter int
	var lenSeriesToDecode = len(filesToDecode)

	for _, fileToDecode := range filesToDecode {
		// decode only .xer files
		if !strings.HasSuffix(fileToDecode, ".xer") {
			continue
		}

		// this is where CONCURRENCY should be added
		serieToAdd, err := decode.Decode(configuration.DatabaseLocalPath, path.Base(fileToDecode))
		if err != nil {
			return nil, fmt.Errorf("download.DecodePartialDatabase(): %s", err.Error())
		}

		seriesDecoded = append(seriesDecoded, serieToAdd)

		serieJSON, err := json.MarshalIndent(serieToAdd, "", "   ")
		if err != nil {
			return nil, fmt.Errorf("download.DecodePartialDatabase(): %s", err.Error())
		}

		fileToWrite := path.Join(configuration.DatabaseLocalPath, serieToAdd.SerieCode+".json")
		//fmt.Printf("File to write: %s\n", fileToWrite)

		err = ioutil.WriteFile(fileToWrite, serieJSON, 0644)
		if err != nil {
			return nil, fmt.Errorf("download.DecodePartialDatabase(): %s", err.Error())
		}

		counter = counter + 1

		fmt.Printf("Decoded: %d\tTotal: %d\r", counter, lenSeriesToDecode)
	}

	fmt.Printf("\nDecoding completed.\n")

	return seriesDecoded, nil

}

// decodes all the .xer files in dbLocalPath and saves the BDSICESeries objects into JSON files
// with the series code as file name. It wraps DecodePartialDatabase() by calling it with
// an array containing all the files in configuration.DatabaseLocalPath
func DecodeFullDatabase(configuration *config.BDSICEConfig) ([]*series.BDSICESerie, error) {

	dbLocalPath := configuration.DatabaseLocalPath

	filesToDecode, err := ioutil.ReadDir(dbLocalPath)
	if err != nil {
		return nil, fmt.Errorf("download.DecodeFullDatabase(): %s", err.Error())
	}

	var filesToDecodePaths []string
	for _, fileToDecode := range filesToDecode {
		if filepath.Ext(fileToDecode.Name()) == ".xer" {
			filesToDecodePaths = append(filesToDecodePaths, fileToDecode.Name())
		}
	}

	seriesDecoded, err := DecodePartialDatabase(configuration, filesToDecodePaths)
	if err != nil {
		return nil, fmt.Errorf("download.DecodeFullDatabase(): error decoding files: %s", err.Error())

	}

	return seriesDecoded, nil
}

// writes the full database BDSICEDatabase object to dbLocalPath in JSON format
// This function could be made redudant if no new series have been added. It would requiere
// checking
func BuildFullDatabase(configuration *config.BDSICEConfig, seriesDecoded []*series.BDSICESerie) error {
	dbLocalPath := configuration.DatabaseLocalPath

	db, err := database.BuildDatabase(seriesDecoded)

	if err != nil {
		return fmt.Errorf("download.BuildFullDatabase(): %s", err.Error())
	}

	/*
		randomSerie := "400000"

		fmt.Printf("Printing a random db element: \nserieCode: %s\ntitle: %s\n", randomSerie, db.Series[randomSerie])
	*/

	dbJSON, err := json.MarshalIndent(db, "", "   ")
	if err != nil {
		return fmt.Errorf("download.BuildFullDatabase(): %s", err.Error())
	}

	fileToWrite := path.Join(dbLocalPath, "db.json")
	//fmt.Printf("File to write: %s\n", fileToWrite)

	err = ioutil.WriteFile(fileToWrite, dbJSON, 0644)
	if err != nil {
		return fmt.Errorf("download.BuildFullDatabase(): %s", err.Error())
	}

	// fmt.Printf("Database built succesfully\n")

	return nil
}

// downloads the full BDSICE database and returns a slice containing the codes of the downloaded series
func DownloadFullDatabase(configuration *config.BDSICEConfig, forceDownload bool) ([]*series.BDSICESerie, error) {
	//	dbLocalPath := path.Join(configuration.DatabaseLocalPath, "db")

	dbLocalPath := configuration.DatabaseLocalPath
	// fmt.Println(dbLocalPath)

	/////////////////
	// first thing is to check if it exists already and if so, prompt unless force true
	if !forceDownload {
		if alreadyDownloadedFullDatabase(configuration, dbLocalPath) { // currently a placeholder: always true
			fmt.Printf("Already existing DB. Do you want to proceed to download anyways?")
			var s string
			fmt.Printf(" (y/n): ")
			fmt.Scan(&s)
			s = strings.TrimSpace(s)
			s = strings.ToLower(s)
			if !(s == "y" || s == "yes") {
				fmt.Printf("Want to proceed with the decoding?")
				var s string
				fmt.Printf(" (y/n): ")
				fmt.Scan(&s)
				s = strings.TrimSpace(s)
				s = strings.ToLower(s)
				if s == "y" || s == "yes" {
					seriesDecoded, err := DecodeFullDatabase(configuration)
					if err != nil {
						return nil, fmt.Errorf("download.DownloadFullDatabase(): %s", err.Error())
					}

					err = BuildFullDatabase(configuration, seriesDecoded)
					if err != nil {
						return nil, fmt.Errorf("download.DownloadFullDatabase(): %s", err.Error())
					}

					return nil, nil
				} else {
					return nil, fmt.Errorf("download.DownloadFullDatabase(): database downloaded but not decoded.")
				}
			} else {
				fmt.Printf("Continuing\n")
			}
		}
	}

	fmt.Printf("Downloading full database to path: %s\n", dbLocalPath)

	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	// FIRST PART:	initial GET request to get session's ids that will be
	//				used after in POST requests.
	client := &http.Client{}

	requestInitialGet, err := http.NewRequest("GET", configuration.DownloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("download.DownloadFullDatabase(): %s.", err.Error())
	}

	// dumping initial GET request
	_, err = httputil.DumpRequestOut(requestInitialGet, true)
	if err != nil {
		return nil, fmt.Errorf("download.DownloadFullDatabase(): %s.", err.Error())
	} else {
		//fmt.Printf("Dumping Initial GET Request\n")
		//fmt.Printf("%s", dump)
	}

	requestInitialGet.Host = "serviciosede.mineco.gob.es"

	headersInitialGetMap := map[string]string{
		"User-Agent":                configuration.UserAgent,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Encoding":           "gzip, deflate",
		"DNT":                       "1",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1"}

	for k, v := range headersInitialGetMap {
		requestInitialGet.Header.Set(k, v)
	}

	responseInitialGet, err := client.Do(requestInitialGet)
	if err != nil {
		return nil, fmt.Errorf("download.DownloadFullDatabase(): %s.", err.Error())
	}
	defer responseInitialGet.Body.Close()

	// dumping initial GET response
	//_, err = httputil.DumpResponse(responseInitialGet, true)
	_, err = httputil.DumpResponse(responseInitialGet, true)

	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	} else {
		//fmt.Printf("Received response to initial GET request. Dumping\n")
		//	fmt.Printf("%s", dump)
	}

	// Extract all the session id parameters from the initial GET response

	root, err := html.Parse(responseInitialGet.Body)

	var eventTargetValue string
	var eventArgumentValue string
	var viewStateValue string
	var viewStateGeneratorValue string
	var eventValidationValue string
	var SessionId string

	eventTarget, ok := getElementById("__EVENTTARGET", root)
	if !ok {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range eventTarget.Attr {
		if attr.Key == "value" {
			eventTargetValue = attr.Val
		}
	}

	eventArgument, _ := getElementById("__EVENTARGUMENT", root)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range eventArgument.Attr {
		if attr.Key == "value" {
			eventArgumentValue = attr.Val
		}
	}

	viewState, _ := getElementById("__VIEWSTATE", root)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range viewState.Attr {
		if attr.Key == "value" {
			viewStateValue = attr.Val
		}
	}

	viewStateGenerator, _ := getElementById("__VIEWSTATEGENERATOR", root)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range viewStateGenerator.Attr {
		if attr.Key == "value" {
			viewStateGeneratorValue = attr.Val
		}
	}

	eventValidation, _ := getElementById("__EVENTVALIDATION", root)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range eventValidation.Attr {
		if attr.Key == "value" {
			eventValidationValue = attr.Val
		}
	}

	for _, cookie := range responseInitialGet.Cookies() {
		if cookie.Name == "ASP.NET_SessionId" {
			SessionId = cookie.Value
		}
	}

	cookieSessionIdValue := fmt.Sprintf("ASP.NET_SessionId=%s", SessionId)
	io.Copy(ioutil.Discard, responseInitialGet.Body)
	responseInitialGet.Body.Close()

	// fmt.Printf("Valores extraídos del primer GET:\neventTarget: %s\neventArgument: %s\nviewState: %s\nviewStateGenerator: %s\neventValidation: %s\ncookieSessionIdValue: %s\n", eventTargetValue, eventArgumentValue, viewStateValue, viewStateGeneratorValue, eventValidationValue, cookieSessionIdValue)

	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	// Second part:	build a post request with the session info that we got
	//				from the initial GET request
	//				Headers and data to be posted need to be built.

	// __EVENTTARGET is an empty string in the answer but must be set to this
	// to download the latest update.

	client = &http.Client{}

	eventTargetValue = "lbutton_BdsiceCompleta"

	dataPostMap := map[string]string{
		"__EVENTTARGET":        eventTargetValue,
		"__EVENTARGUMENT":      eventArgumentValue,
		"__VIEWSTATE":          viewStateValue,
		"__VIEWSTATEGENERATOR": viewStateGeneratorValue,
		"__EVENTVALIDATION":    eventValidationValue}

	// as urlencoded form
	postData := url.Values{}
	for k, v := range dataPostMap {
		postData.Set(k, v)
	}

	postDataEncoded := postData.Encode()

	requestPost, err := http.NewRequest("POST", configuration.DownloadURL, strings.NewReader(postDataEncoded))
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	headersPostMap := map[string]string{
		"User-Agent":                configuration.UserAgent,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Encoding":           "gzip, deflate",
		"Connection":                "keep-alive",
		"Content-Type":              "application/x-www-form-urlencoded",
		"Content-Length":            strconv.Itoa(len(postDataEncoded)),
		"Origin":                    "http://serviciosede.mineco.gob.es",
		"DNT":                       "1",
		"Referer":                   configuration.DownloadURL,
		"Cookie":                    cookieSessionIdValue,
		"Upgrade-Insecure-Requests": "1"}

	for k, v := range headersPostMap {
		requestPost.Header.Set(k, v)
		// fmt.Printf("%s : %s\n", k, v)
	}

	requestPost.Host = "serviciosede.mineco.gob.es"

	//// Actual POST request
	//fmt.Println("\nPerforming POST request")
	responsePost, err := client.Do(requestPost)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	var zipFilePath string

	// download the database contained in a .zip file from POST response body
	zipFilePath, _, err = downloadWithCounter(configuration, responsePost)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	// extract zip file into db folder
	fmt.Printf("Unzipping database file...\n")
	_, err = unzipFile(zipFilePath, dbLocalPath, true)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	fmt.Printf("Decoding .xer files into .json...\n")
	// decode .xer files extracted into .json files
	seriesDecoded, err := DecodeFullDatabase(configuration)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	// populate BDSICEDatabase with all the series available. BuildFullDatabase takes
	// a slice of pointers to series.BDSICESerie objects.
	err = BuildFullDatabase(configuration, seriesDecoded)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	return seriesDecoded, nil
}

// checks for updates and download to local path
// TODO:
//			- Add updated series to json database
// returns:
// 	- slice containing extracted files
// 	- error
func Update(configuration *config.BDSICEConfig, forceUpdate bool) ([]*series.BDSICESerie, error) {

	//	updateLocalPath := path.Join(configuration.DatabaseLocalPath, "updates")
	updateLocalPath := configuration.DatabaseLocalPath
	fmt.Printf("Updating to path: %s\n", updateLocalPath)

	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	// FIRST PART:	initial GET request to get session's ids that will be
	//				used after in POST requests.
	client := &http.Client{}

	requestInitialGet, err := http.NewRequest("GET", configuration.UpdateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Update(): %s.", err.Error())
	}

	// dumping initial GET request
	_, err = httputil.DumpRequestOut(requestInitialGet, true)
	if err != nil {
		return nil, fmt.Errorf("Update(): %s.", err.Error())
	} else {
		//		fmt.Printf("Initial GET Request\n")
		//		fmt.Printf("%s", dump)
	}

	requestInitialGet.Host = "serviciosede.mineco.gob.es"

	headersInitialGetMap := map[string]string{
		"User-Agent":                configuration.UserAgent,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Encoding":           "gzip, deflate",
		"DNT":                       "1",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1"}

	for k, v := range headersInitialGetMap {
		requestInitialGet.Header.Set(k, v)
	}

	responseInitialGet, err := client.Do(requestInitialGet)
	if err != nil {
		return nil, fmt.Errorf("Update(): %s.", err.Error())
	}
	defer responseInitialGet.Body.Close()

	/* dumping initial GET response
	_, err = httputil.DumpResponse(responseInitialGet, true)
	if err != nil {
		return nil, fmt.Errorf("(): element dg_Actualizaiones_ctl2_boton could not be found.", err.Error())
	} else {
		//	fmt.Printf("Initial GET Request response\n")
		//fmt.Printf("%s", dump)
	}
	*/

	// Extract all the session id parameters from the initial GET response

	root, err := html.Parse(responseInitialGet.Body)

	/////////////////
	// check if update exists already
	if !forceUpdate {
		lastUpdateLink, ok := getElementById("dg_Actualizaciones__ctl2_boton", root)
		if !ok {
			return nil, fmt.Errorf("Update(): %s.", err.Error())
		}

		lastUpdateText := lastUpdateLink.FirstChild.Data

		var day, month, year int

		fmt.Sscanf(lastUpdateText, "Series actualizadas el día %d del %d de %d",
			&day, &month, &year)

		if alreadyDownloadedUpdate(updateLocalPath, day, month, year) {
			fmt.Printf("Update for %d-%d-%d has already been downloaded.\n",
				day, month, year)
			return nil, fmt.Errorf("Update(): Latest available update has been already downloaded.")
		}
	}

	// We have checked for updates already downloaded. We need to download the
	// newest update that we don't have. First step is to extract the id's of
	// the session that will be used to create a POST request that will
	// give access to the file.

	var eventTargetValue string
	var eventArgumentValue string
	var viewStateValue string
	var viewStateGeneratorValue string
	var eventValidationValue string
	var SessionId string

	eventTarget, ok := getElementById("__EVENTTARGET", root)
	if !ok {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range eventTarget.Attr {
		if attr.Key == "value" {
			eventTargetValue = attr.Val
		}
	}

	eventArgument, _ := getElementById("__EVENTARGUMENT", root)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range eventArgument.Attr {
		if attr.Key == "value" {
			eventArgumentValue = attr.Val
		}
	}

	viewState, _ := getElementById("__VIEWSTATE", root)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range viewState.Attr {
		if attr.Key == "value" {
			viewStateValue = attr.Val
		}
	}

	viewStateGenerator, _ := getElementById("__VIEWSTATEGENERATOR", root)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range viewStateGenerator.Attr {
		if attr.Key == "value" {
			viewStateGeneratorValue = attr.Val
		}
	}

	eventValidation, _ := getElementById("__EVENTVALIDATION", root)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	for _, attr := range eventValidation.Attr {
		if attr.Key == "value" {
			eventValidationValue = attr.Val
		}
	}

	for _, cookie := range responseInitialGet.Cookies() {
		if cookie.Name == "ASP.NET_SessionId" {
			SessionId = cookie.Value
		}
	}

	cookieSessionIdValue := fmt.Sprintf("ASP.NET_SessionId=%s", SessionId)
	io.Copy(ioutil.Discard, responseInitialGet.Body)
	responseInitialGet.Body.Close()

	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////
	// Second part:	build a post request with the session info that we got
	//				from the initial GET request
	//				Headers and data to be posted need to be built.

	// __EVENTTARGET is an empty string in the answer but must be set to this
	// to download the latest update.

	client = &http.Client{}

	eventTargetValue = "dg_Actualizaciones$_ctl2$boton"

	dataPostMap := map[string]string{
		"__EVENTTARGET":        eventTargetValue,
		"__EVENTARGUMENT":      eventArgumentValue,
		"__VIEWSTATE":          viewStateValue,
		"__VIEWSTATEGENERATOR": viewStateGeneratorValue,
		"__EVENTVALIDATION":    eventValidationValue}

	// as urlencoded form
	postData := url.Values{}
	for k, v := range dataPostMap {
		postData.Set(k, v)
	}

	postDataEncoded := postData.Encode()

	requestPost, err := http.NewRequest("POST", configuration.UpdateURL, strings.NewReader(postDataEncoded))
	if err != nil {
		//log.Fatalln(err)
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	headersPostMap := map[string]string{
		"User-Agent":      configuration.UserAgent,
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Encoding": "gzip, deflate",
		"Connection":      "keep-alive",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Content-Length":  strconv.Itoa(len(postDataEncoded)),
		"Origin":          "http://serviciosede.mineco.gob.es",
		"DNT":             "1",
		//"Referer":                   "http://serviciosede.mineco.gob.es/Indeco/BDSICE/UltimasActualizaciones_new.aspx",
		// "Referer":                   configuration.UpdateURL,
		"Referer":                   "http://serviciosede.mineco.gob.es/Indeco/BDSICE/Ultimasactualizaciones_new.aspx",
		"Cookie":                    cookieSessionIdValue,
		"Upgrade-Insecure-Requests": "1"}

	for k, v := range headersPostMap {
		requestPost.Header.Set(k, v)
		//fmt.Printf("%s : %s\n", k, v)
	}

	requestPost.Host = "serviciosede.mineco.gob.es"

	responsePost, err := client.Do(requestPost)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	// POST must return 200 OK. But it is not a sufficient condition. DownloadWithCounter will further
	// check for a Content-Disposition header in the response to the final GET
	if responsePost.StatusCode == 200 {
		//fmt.Println("\nResponse to POST is 200 OK")
	} else {
		//log.Fatalln("Response to POST was not 200 OK. Has the website changed?")
		return nil, fmt.Errorf("POST request did not return 200 OK")
	}

	/////////////////////////////////////
	// THIRD PART: Second GET: using the link returned to the POST request, an .zip should be returned
	/////////////////////////////////////

	//requestGetSecond, err := http.NewRequest("GET", configuration.UpdateURL, strings.NewReader(postDataEncoded))

	client = &http.Client{}

	requestFinalGet, err := http.NewRequest("GET", "http://serviciosede.mineco.gob.es/Indeco/DescargaArchivo.aspx?estadisticas=True&tipo=1", nil)
	if err != nil {
		// log.Fatalln(err)
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	requestFinalGet.Host = "serviciosede.mineco.gob.es"

	headersFinalGetMap := map[string]string{
		"User-Agent":                configuration.UserAgent,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Encoding":           "gzip, deflate",
		"Connection":                "keep-alive",
		"Origin":                    "http://serviciosede.mineco.gob.es",
		"DNT":                       "1",
		"Referer":                   "http://serviciosede.mineco.gob.es/Indeco/BDSICE/Ultimasactualizaciones_new.aspx",
		"Cookie":                    cookieSessionIdValue,
		"Upgrade-Insecure-Requests": "1"}

	for k, v := range headersFinalGetMap {
		//fmt.Printf("Setting --- %s\t%s\n", k, v)
		requestFinalGet.Header.Set(k, v)
	}

	// dumping final GET request
	/* dump, err = httputil.DumpRequestOut(requestFinalGet, true)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("\n\n\n\nFinal GET Request\n")
		fmt.Printf("%s", dump)
	}
	*/

	responseFinalGet, err := client.Do(requestFinalGet)
	if err != nil {
		//log.Fatal(err)
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}
	defer responseFinalGet.Body.Close()

	// dumping final GET response
	/*
		dump, err = httputil.DumpResponse(responseFinalGet, true)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("%s\n", dump)
		}
	*/

	// Printing headers of the final GET, without gibberish body (since it should be a zip file)
	/*
		fmt.Println("Headers of the final GET")
		for k, v := range responseFinalGet.Header {
			fmt.Printf("%s:\t%s\n", k, v)
		}
	*/

	var zipFilePath string

	// download from GET response body into .zip file
	zipFilePath, _, err = downloadWithCounter(configuration, responseFinalGet)

	// unzip files into update folder
	extractedFiles, err := unzipFile(zipFilePath, updateLocalPath, false)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	/* We might want to print all the series to be updated.

	db, err := database.LoadDatabase(configuration)
	if err != nil {
		return nil, fmt.Errorf("LoadDatabase(): %s", err.Error())
	}

	fmt.Println("Series to update:")
	for _, extractedFile := range extractedFiles {
		// fmt.Printf("extractedFile: %s\n", extractedFile)
		extension := filepath.Ext(extractedFile)
		if extension == ".xer" {
			baseName := filepath.Base(extractedFile)
			serie := strings.TrimSuffix(baseName, extension)
			// fmt.Printf("%s\t\t%s\n", serie, db.Series[serie])
		}

	}
	*/

	fmt.Println("Moving updated series to database folder...")
	var copiedFiles []string
	for _, extractedFile := range extractedFiles {
		input, err := ioutil.ReadFile(extractedFile)
		if err != nil {
			return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
		}

		base := filepath.Base(extractedFile)
		output := filepath.Join(configuration.DatabaseLocalPath, base)
		//fmt.Printf("Copying to %s\n", output)

		err = ioutil.WriteFile(output, input, 0644)
		if err != nil {
			return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
		}

		copiedFiles = append(copiedFiles, output)

	}

	// Firstly, let's decode the series that have been updated (and only them)
	fmt.Printf("Decoding database...\n")
	seriesDecoded, err := DecodePartialDatabase(configuration, copiedFiles)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	// Secondly, rebuild database file (althought it probably shouldn't, but in case a new serie has been added)
	fmt.Printf("Rebuilding database...\n")
	err = BuildFullDatabase(configuration, seriesDecoded)
	if err != nil {
		return nil, fmt.Errorf("DownloadFullDatabase(): %s.", err.Error())
	}

	return seriesDecoded, nil
}

// download latest available "coyuntura" bulletin
func Bulletin(configuration *config.BDSICEConfig) error {
	fmt.Printf("Bulletin download not yet implemented.\n")
	return nil
}
