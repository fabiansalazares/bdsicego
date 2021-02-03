// config package to read configuration files
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v2"
)

const DefaultConfigFilePath = ".config/bdsicego/config.yml"
const DefaultConfigFileContent = `
updateurl: "http://serviciosede.mineco.gob.es/Indeco/BDSICE/Ultimasactualizaciones_new.aspx"
bulletinurl: "http://serviciosede.mineco.gob.es/indeco/"
downloadurl: "http://serviciosede.mineco.gob.es/Indeco/BDSICE/HomeBDSICE.aspx"
useragent: "Mozilla/5.0 (Windows NT 10.0; rv:68.0) Gecko/20100101 Firefox/68.0"
debug: true
`

/* Additional fields to add at runtime:
path: "<PATH TO DATA FOLDER>"
dblocalpath: "<PATH TO FOLDER WITHIN PATH WITH .XER AND .JSON FILES>"
plotviewer: "<IMAGE VIEWER COMMAND>"
*/

// An additional field plotviewer: "VIEWER" is added at runtime, VIEWER being an archicture-dependant value

type BDSICEConfig struct {
	DataLocalPath     string `yaml:"path"`
	DatabaseLocalPath string `yaml:"dblocalpath"`
	UpdateLocalPath   string `yaml:"updatelocalpath"`
	UpdateURL         string `yaml:"updateurl"`
	BulletinURL       string `yaml:"bulletinurl"`
	DownloadURL       string `yaml:"downloadurl"`
	UserAgent         string `yaml:"useragent"`
	Debug             bool   `yaml:"debug"`
	PlotViewer        string `yaml:"plotviewer"`
}

// returns a hard-coded and architecture-dependant path to the config file.
func GetDefaultConfigFilePath() (string, error) {
	var filePath string

	configDir := configdir.LocalConfig("bdsicego")
	err := configdir.MakePath(configDir)

	if err != nil {
		return "", fmt.Errorf("config.GetConfig(): %s", err.Error())
	}

	filePath = filepath.Join(configDir, "config.yml")

	// return path.Join(home, ".local/share/bdsicego/config.yml")
	return filePath, nil
}

//func GetConfig(filePath string) *BDSICEConfig {
func GetConfig() (*BDSICEConfig, error) {
	var configuration BDSICEConfig
	var filePath string

	configDir := configdir.LocalConfig("bdsicego")
	err := configdir.MakePath(configDir)

	if err != nil {
		return nil, fmt.Errorf("config.GetConfig(): %s", err.Error())
	}

	filePath = filepath.Join(configDir, "config.yml")

	// fmt.Printf("filePath: %s\n", filePath)

	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		var DefaultRuntimeString string
		var input string
		fmt.Printf("No configuration file found. Would you like to create one with default values? ")
		fmt.Printf(" (y/n) ")
		fmt.Scan(&input)
		input = strings.TrimSpace(input)
		input = strings.ToLower(input)
		if input == "y" || input == "yes" || input == "" {
			var DefaultString string

			fmt.Printf("Creating a new file at default configuration directory: %s...\n", filePath)

			if strings.Contains(runtime.GOOS, "linux") {
				DefaultRuntimeString = `plotviewer: "feh"`
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "path: " + "\"" + filepath.Join(configDir) + "\"" + "\n"
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "dblocalpath: " + "\"" + filepath.Join(configDir) + "/db" + "\"" + "\n"
				DefaultString = DefaultConfigFileContent + DefaultRuntimeString
			} else if strings.Contains(runtime.GOOS, "windows") {
				DefaultRuntimeString = `plotviewer: "explorer"`
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "path: " + "\"" + filepath.Join(configDir) + "\"" + "\n"
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "dblocalpath: " + "\"" + filepath.Join(configDir) + "\\db" + "\"" + "\n"
				DefaultString = DefaultConfigFileContent + DefaultRuntimeString
				DefaultString = strings.Replace(DefaultString, "\\", "\\\\", -1)
			} else if strings.Contains(runtime.GOOS, "darwin") {
				DefaultRuntimeString = `plotviewer: "open"`
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "path: " + "\"" + filepath.Join(configDir) + "\"" + "\n"
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "dblocalpath: " + "\"" + filepath.Join(configDir) + "/db" + "\"" + "\n"
				DefaultString = DefaultConfigFileContent + DefaultRuntimeString
			} else {
				fmt.Printf("Operating system could not be detected. Asuming unix-like system and feh viewer available.")
				DefaultRuntimeString = `plotviewer: "feh"`
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "path: " + "\"" + filepath.Join(configDir) + "\"" + "\n"
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "dblocalpath: " + "\"" + filepath.Join(configDir) + "/db" + "\"" + "\n"
				DefaultString = DefaultConfigFileContent + DefaultRuntimeString
			}

			// to hold the full config file to be written

			err = ioutil.WriteFile(filePath, []byte(DefaultString), 0644)
			if err != nil {
				return nil, fmt.Errorf("config.GetConfig(): %s", err.Error())
			}

			// do not return a BDSICEConfig object yet because we will read it from the file that we just wrote
		} else {
			fmt.Printf("No configuration file was created. Proceeding with default hardcoded values. ")

			// a list of available os/platform can be obtained with command $ go tool dist list"
			var DefaultString string

			if strings.Contains(runtime.GOOS, "linux") {
				DefaultRuntimeString = `plotviewer: "feh"`
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "path: " + "\"" + filepath.Join(configDir) + "\"" + "\n"
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "dblocalpath: " + "\"" + filepath.Join(configDir) + "/db" + "\"" + "\n"
				DefaultString = DefaultConfigFileContent + DefaultRuntimeString
			} else if strings.Contains(runtime.GOOS, "windows") {
				DefaultRuntimeString = `plotviewer: "explorer"`
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "path: " + "\"" + filepath.Join(configDir) + "\"" + "\n"
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "dblocalpath: " + "\"" + filepath.Join(configDir) + "\\db" + "\"" + "\n"
				DefaultString = DefaultConfigFileContent + DefaultRuntimeString
			} else if strings.Contains(runtime.GOOS, "darwin") {
				DefaultRuntimeString = `plotviewer: "open"`
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "path: " + "\"" + filepath.Join(configDir) + "\"" + "\n"
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "dblocalpath: " + "\"" + filepath.Join(configDir) + "/db" + "\"" + "\n"
				DefaultString = DefaultConfigFileContent + DefaultRuntimeString
				DefaultString = strings.Replace(DefaultString, "\\u", "\\\\u", -1)
			} else {
				fmt.Printf("Operating system could not be detected. Asuming unix-like system.")
				DefaultRuntimeString = `plotviewer: "feh"`
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "path: " + "\"" + filepath.Join(configDir) + "\"" + "\n"
				DefaultRuntimeString = DefaultRuntimeString + "\n" + "dblocalpath: " + "\"" + filepath.Join(configDir) + "/db" + "\"" + "\n"
				DefaultString = DefaultConfigFileContent + DefaultRuntimeString
			}

			decoder := yaml.NewDecoder(strings.NewReader(DefaultString))
			err = decoder.Decode(&configuration)

			if err != nil {
				return nil, fmt.Errorf("config.GetConfig(): %s", err.Error())
			}

			return &configuration, nil
		}
	}

	fileHandler, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("config.GetConfig(): %s", err.Error())
	}
	defer fileHandler.Close()

	decoder := yaml.NewDecoder(fileHandler)
	err = decoder.Decode(&configuration)

	if err != nil {
		return nil, fmt.Errorf("config.GetConfig(): %s", err.Error())
	}

	// make sure that DataLocalPath/db/ exists because we might want to store database there
	err = configdir.MakePath(configuration.DatabaseLocalPath)
	if err != nil {
		return nil, fmt.Errorf("config.GetConfig(): %s", err.Error())
	}

	return &configuration, nil
}
