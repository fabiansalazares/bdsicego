// bdsicego shows, compares, plots, updates and downloads BDSICE data.
package main

import (
	"math"
	"math/rand"
	"os/exec"
	"time"

	// "bdsice/decode"

	"github.com/fabiansalazares/bdsicego/database"
	"github.com/fabiansalazares/bdsicego/download"
	"github.com/fabiansalazares/bdsicego/internal/config"
	"github.com/fabiansalazares/bdsicego/internal/version"
	"github.com/fabiansalazares/bdsicego/plot"
	"github.com/fabiansalazares/bdsicego/series"

	// econseries "econdata/series"
	//	econseries "fabiansalazares/bdsicego/series"

	"strings"

	// "fabiansalazares/bdsicego/bdsice/series"
	"fmt"
	"log"
	"os"

	// to print nicely formatted tables for show and compare commands
	"github.com/c-bata/go-prompt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const helpMessage = `
	Commands:


	h | help  			prints this message
	v | version  			prints version
	e | setup 			prints the current configuration parameters
	d | download (force) 		downloads the full database from the BDSICE website
	u | update 			downloads the most recent update from the BDSICE website
	b | bulletin 			downloads the most recent coyuntura bulletin from BDSICE website
	i | info [codes]		prints information about the given codes
	s | search [terms] 		searches the terms in the local BDSICE database
	w | show [%%] [codes] 		prints a summary of the specified codes or matched codes if "%%"
	c | compare [codes] 		compares the series given
	p | plot [%%] [sep] [codes]    	plots the series given. "%%" includes codes matched from search commands
	r | random 			shows a randomly chosen serie
`

// custom type holding arguments to a search command
type searchArgs struct {
	terms []string
}

// custom type holding arguments to a show command
type showArgs struct {
	active bool
	codes  []string
}

// custom type holding arguments to a compare command
type compareArgs struct {
	active bool
	codes  []string
}

// custom type holding arguments to a plot command
type plotArgs struct {
	active   bool
	separate bool
	save     bool
	codes    []string
}

type infoArgs struct {
	active bool
	codes  []string
}

// custom type to hold a representation of the commands to execute.
type argsStruct struct {
	search          []searchArgs
	searchToInfo    bool
	searchToShow    bool
	searchToCompare bool
	searchToPlot    bool
	show            showArgs
	compare         compareArgs
	plot            plotArgs
	info            infoArgs
	verbose         bool
}

func promptCompleter(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "quit", Description: "quit the current session"},
		{Text: "help", Description: "display available commands and options"},
		{Text: "version", Description: "print current version"},
		{Text: "setup", Description: "display current setup"},
		{Text: "info", Description: "show basic information about specified serie(s)"},
		{Text: "download", Description: "download the full database"},
		{Text: "update", Description: "download the latest update"},
		{Text: "bulletin", Description: "download and display the latest bulletin"},
		{Text: "info", Description: "display basic information about specified serie(s)"},
		{Text: "search", Description: "search for series whose codes or titles contain given search terms"},
		{Text: "show", Description: "show the specified serie(s)"},
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)

}

// prints usage information and general help
func helpCommand(configuration *config.BDSICEConfig) {
	fmt.Printf("%s %s \t queries and shows BDSICE database info.\n", version.CmdName, version.CmdVersion)
	fmt.Printf(helpMessage)
}

// prints current version
func versionCommand(configuration *config.BDSICEConfig) {

	fmt.Printf("%s\n", version.CmdVersion)
	return
}

// prints basic information for the given code series
func infoCommand(configuration *config.BDSICEConfig, commandArgs *argsStruct) {
	for _, code := range commandArgs.info.codes {
		serie, err := series.Load(configuration, code)
		if err != nil {
			fmt.Printf("Serie code %s does not exist in the BDSICE database.\n", code)
			continue
		}

		fmt.Printf(`
BDSICE Serie %s -- %s
Range: %s to %s
Number of observations: %d
Source: %s
Units: %s
Number of decimals: %d
Frequency: %d
`, serie.SerieCode,
			serie.Title,
			serie.Start.String(),
			serie.End.String(),
			serie.NumberOfObservations,
			serie.Source,
			serie.Units,
			serie.Decimals,
			serie.Frequency,
		)

	}

	return
}

// shows current configuration and sets configuration values
func setupCommand(configuration *config.BDSICEConfig) {

	//config, err := config.GetConfig()
	//if err != nil {
	//	log.Fatal(err)
	//}

	fmt.Printf("DataLocalPath: %s\nDatabaseLocalPath: %s\nUpdateURL: %s\nBulletinURL: %s\nDownloadURL: %s\nUserAgent: %s\nDebug: %v\nPlotViewer: %s\n",
		configuration.DataLocalPath,
		configuration.DatabaseLocalPath,
		configuration.UpdateURL,
		configuration.BulletinURL,
		configuration.DownloadURL,
		configuration.UserAgent,
		configuration.Debug,
		configuration.PlotViewer)
	return
}

// downloads the full database and process .xer files into .json loadable ones
func downloadCommand(configuration *config.BDSICEConfig, force bool) {

	// _, err := download.DownloadFullDatabase(configuration, false)
	_, err := download.DownloadFullDatabase(configuration, force)

	if err != nil {
		log.Fatal(err)
	}

	return
}

// updates database in path DatabaseLocalPath config parameter
// TODO:
//		- Implement force modifier to update command that forces the update even if it has already been deownloaded.
func updateCommand(configuration *config.BDSICEConfig) {

	_, err := download.Update(configuration, false)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Sprintln()
	return

}

// TODO checks for the latest bulletin and download if available
func bulletinCommand(configuration *config.BDSICEConfig) {

	err := download.Bulletin(configuration)

	if err != nil {
		log.Fatal(err)
	}

	return

}

// searchs for the given terms. If commands show, plot or info contain % as an arg, the series returned by
// search will also by taken as arguments to those commands
//func searchCommand(configuration *config.BDSICEConfig, commandArgs *argsStruct) (map[string]string, error) {
func searchCommand(configuration *config.BDSICEConfig, commandArgs *argsStruct, resultsStack map[string]string) error {

	if len(commandArgs.search) == 0 {
		return nil
	}

	//var resultsToReturn = map[string]string{}

	// clean results from a previous search, if a map reference was provided as argument
	// instead of pointing resultsStack to a new map[string]string object, we need to remove
	// all previous k,v pair. This is so because we need to keep resultsStack pointer to the same
	// object that we passed as reference and if we allocated a new map with make, we would be
	// changing the reference of resultsStack altogether, and we would not be adding the results
	// as k,v to the original object passed as reference.
	if resultsStack != nil {
		//fmt.Println("Cleaning")
		// resultsStack = make(map[string]string)
		for k, _ := range resultsStack {
			delete(resultsStack, k)
		}
	}

	// loop through each call to searchCommand, then perform a search for each call
	for _, searchCall := range commandArgs.search {

		if len(searchCall.terms) == 0 {

			fmt.Printf("searchCommand: nothing to search for.\n")
			continue
		}

		// load database into 'db' variable
		db, err := database.LoadDatabase(configuration)
		if err != nil {
			return fmt.Errorf("searchCommand(): %s", err.Error())
		}

		// perform search using terms as variadic arguments
		resultsDatabaseSeries, err := db.Search(searchCall.terms...)
		if err != nil {
			return fmt.Errorf("searchCommand(): %s", err.Error())
		}

		// show results
		for code, title := range resultsDatabaseSeries {

			fmt.Printf("%s\t\t%s\n", code, title)

			// load results into result stack, if a map reference was provided as argument
			if resultsStack != nil {
				//fmt.Printf("Loading results: %s %s\n", code, title)
				resultsStack[code] = title
			}

			if commandArgs.searchToInfo {
				commandArgs.info.codes = append(commandArgs.info.codes, code)
			}

			if commandArgs.searchToCompare {
				commandArgs.compare.codes = append(commandArgs.compare.codes, code)
			}

			if commandArgs.searchToPlot {
				commandArgs.plot.codes = append(commandArgs.plot.codes, code)
			}

			if commandArgs.searchToShow {
				commandArgs.show.codes = append(commandArgs.show.codes, code)
			}
		}
	}

	return nil

}

// displays a table containing the data in the given serie, including max, min, average and period growth
func showSerie(s *series.BDSICESerie) {
	var yoy float64

	t := table.NewWriter()
	//t.SetColumnPainter(colorFunc)
	t.SetOutputMirror(os.Stdout)
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}

	// Serie code and title header
	t.AppendHeader(table.Row{s.GetCode(), s.GetCode()}, rowConfigAutoMerge)
	t.AppendHeader(table.Row{s.GetTitle()}, rowConfigAutoMerge)

	// Data values
	t.AppendHeader(table.Row{"Period", strings.TrimSpace(s.Units), "Growth"})

	// implicit function that gets called for every row in the third column
	// it checks for a - prefixed to the cell's content and if it finds it, it colors the
	// cell accordingly  to denote negative growth in the corresponding period
	checkSignYoY := text.Transformer(func(val interface{}) string {
		if strings.HasPrefix(val.(string), "-") {
			// YoY string contains a negative value
			return text.FgHiRed.Sprint(val)
		} else {

			return text.FgGreen.Sprint(val)
		}
	})

	cconfigs := []table.ColumnConfig{
		{Name: "Period", Align: text.AlignLeft, AlignHeader: text.AlignCenter, WidthMin: 6, WidthMax: 25},
		{Name: "Value", Align: text.AlignLeft, WidthMin: 3, WidthMax: 25},
		{Name: "Growth", Align: text.AlignRight, Transformer: checkSignYoY}}

	t.SetColumnConfigs(cconfigs)
	//		t.SetColumnConfigs([]table.ColumnConfig{{}, {}, {}})

	//var rowConfigYoYColor table.RowConfig

	// if units is not percentage
	if !strings.Contains(s.Units, "PORCENTAJE") {
		for i := 0; i < len(s.Observations.Dates); i++ {
			if i > 0 {
				// Calculate YoY growth if we are looping past the first observation
				yoy = ((s.Observations.Values[i] - s.Observations.Values[i-1]) / s.Observations.Values[i-1]) * 100
				timeString := fmt.Sprintf("%d:%d", s.Observations.Dates[i].Year(), s.Observations.Dates[i].Month())
				t.AppendRow(table.Row{timeString, fmt.Sprintf("%10.2f", s.Observations.Values[i]), fmt.Sprintf("%.1f %%", yoy)})

			} else {
				// do not calculate
				yoy = math.NaN()
				timeString := fmt.Sprintf("%d:%d", s.Observations.Dates[i].Year(), s.Observations.Dates[i].Month())
				t.AppendRow(table.Row{timeString, fmt.Sprintf("%10.2f", s.Observations.Values[i]), ""})
			}

		}
	} else {
		for i := 0; i < len(s.Observations.Dates); i++ {
			if i > 0 {
				// Calculate YoY growth if we are looping past the first observation
				yoy = (s.Observations.Values[i] - s.Observations.Values[i-1])
				timeString := fmt.Sprintf("%d:%d", s.Observations.Dates[i].Year(), s.Observations.Dates[i].Month())
				t.AppendRow(table.Row{timeString, fmt.Sprintf("%10.2f", s.Observations.Values[i]), fmt.Sprintf("%.1f %%", yoy)})

			} else {
				// do not calculate
				yoy = math.NaN()
				timeString := fmt.Sprintf("%d:%d", s.Observations.Dates[i].Year(), s.Observations.Dates[i].Month())
				t.AppendRow(table.Row{timeString, fmt.Sprintf("%10.2f", s.Observations.Values[i]), ""})
			}

		}

	}

	// final rows with max, min and average for the serie
	t.AppendSeparator()

	// START PERIOD
	startPeriod := s.Observations.Dates[0]
	startPeriodString := fmt.Sprintf("%d:%d", startPeriod.Year(), startPeriod.Month())
	t.AppendRow([]interface{}{"Start", startPeriodString, ""})

	// END PERIOD
	endPeriod := s.Observations.Dates[len(s.Observations.Dates)-1]
	endPeriodString := fmt.Sprintf("%d:%d", endPeriod.Year(), endPeriod.Month())
	t.AppendRow([]interface{}{"End", endPeriodString, ""})

	// MAX VALUE
	maxValue, maxTime := s.Max()
	maxTimeString := fmt.Sprintf("%d:%d", maxTime.Year(), maxTime.Month())
	t.AppendRow([]interface{}{"Max", maxTimeString, fmt.Sprintf("%.1f", maxValue)})

	// MIN VALUE
	minValue, minTime := s.Min()
	minTimeString := fmt.Sprintf("%d:%d", minTime.Year(), minTime.Month())
	t.AppendRow([]interface{}{"Min", minTimeString, fmt.Sprintf("%.1f", minValue)})

	// AVERAGE
	average := s.Average()
	t.AppendRow([]interface{}{"Average", "", fmt.Sprintf("%.1f", average)})

	// Serie code and title header
	t.AppendFooter(table.Row{s.GetCode(), s.GetCode()}, rowConfigAutoMerge)
	t.AppendFooter(table.Row{s.GetTitle(), s.GetTitle()}, rowConfigAutoMerge)

	t.SetStyle(table.StyleColoredBright) //GreenWhiteOnBlack)
	t.Render()

}

// TODO prints the data and main dimensions of the given codes
func showCommand(configuration *config.BDSICEConfig, commandArgs *argsStruct) {
	if !commandArgs.show.active {
		return
	}

	for _, code := range commandArgs.show.codes {
		s, err := series.Load(configuration, code)
		if err != nil {
			fmt.Printf("Show: %s could not be loaded, skipping...\n", err.Error())
			continue
		}
		showSerie(s)
	}

	return
}

// TODO compares the given serie codes
func compareCommand(configuration *config.BDSICEConfig, commandArgs *argsStruct) {
	if !commandArgs.compare.active {
		return
	}

	for i, code := range commandArgs.compare.codes {
		fmt.Printf("term %d: %s\n", i, code)
	}

	return
}

// TODO plots the given serie codes in a single plot, or separatedly if so specified with "separate" command modifier
func plotCommand(configuration *config.BDSICEConfig, commandArgs *argsStruct) {
	if len(commandArgs.plot.codes) == 0 {
		// nothing to plot if codes are not larger than 0!
		//fmt.Printf("plotCommand: nothing to plot.\n")
		return
	}

	// we are creating an array of EconSerie, the interface type and not directly an array of BDSICESeries
	// Apparently, arrays of types that implement an interface cannot be passed as arguments to a function
	// that takes an array of the implemented interface. plot.Plot() takes []series.EconSerie, so we could not
	// pass a []series.BSDICESerie object and expect golang to convert it because apparently it is computationally
	// expensive. See: https://www.reddit.com/r/golang/comments/3gtg3i/passing_slice_of_values_as_slice_of_interfaces/
	// so the only way is to create an array of econdata/series.EconSerie and convert each BDSICESerie object
	// before appending it.
	//var seriesToPlot []series.EconSerie
	var seriesToPlot []series.BDSICESerie

	if commandArgs.plot.separate {
		// we will plot each serie to a separate file
		for _, code := range commandArgs.plot.codes {
			serieToPlot, err := series.Load(configuration, code)
			if err != nil {
				fmt.Printf("Serie %s could not be loaded. It will not be plotted.\n", code)
				continue
			}
			//tmpFile, err := plot.Plot(series.EconSerie(*(serieToPlot)))
			tmpFile, err := plot.Plot(*(serieToPlot))

			if err != nil {
				fmt.Printf("plotCommand: an error ocurred while plotting: %s", err.Error())
			}

			cmd := exec.Command(configuration.PlotViewer, tmpFile)

			err = cmd.Start()
			if err != nil {
				fmt.Printf("plotCommand: an error ocurred while executing kitten: %s", err.Error())
			}
		}
	} else {
		// joint plotting by default
		for i, code := range commandArgs.plot.codes {
			serieToPlot, err := series.Load(configuration, code)
			if err != nil {
				fmt.Printf("Serie %s could not be loaded. It will not be plotted.\n", code)
				continue
			}
			//seriesToPlot = append(seriesToPlot, series.EconSerie(*(serieToPlot)))
			seriesToPlot = append(seriesToPlot, *(serieToPlot))
			fmt.Printf("code %d: %s\n", i, code)
		}

		if len(seriesToPlot) > 0 {
			tmpFile, err := plot.Plot(seriesToPlot...)
			if err != nil {
				fmt.Printf("plotCommand: an error ocurred while plotting: %s", err.Error())
			}

			// cmd := exec.Command("kitty", "@ kitten icat", tmpFile)
			//cmd := exec.Command("kitty", "+kitten icat", tmpFile)
			cmd := exec.Command(configuration.PlotViewer, tmpFile)

			err = cmd.Start()
			if err != nil {
				fmt.Printf("plotCommand: an error ocurred while executing kitten: %s", err.Error())
			}
		}

		// separate plots
		return
	}
	return
}

// extracts a random serie code from the database and shows it
func randomCommand(configuration *config.BDSICEConfig) error {
	rand.Seed(time.Now().UTC().UnixNano())
	fmt.Println("Random serie")

	db, err := database.LoadDatabase(configuration)
	if err != nil {
		return fmt.Errorf("randomCommand(): %s", err.Error())
	}
	/*
		serieCodes := make([]string, 0, len(db.Series))

		for k := range db.Series {
			serieCodes = append(serieCodes, k)
		}
	*/
	randomCode := db.Codes[rand.Int()%len(db.Series)]

	s, err := series.Load(configuration, randomCode)
	if err != nil {
		fmt.Printf("Show: %s could not be loaded, skipping...\n", err.Error())
		return fmt.Errorf("randomCommand(): %s", err.Error())
	}

	showSerie(s)

	return nil

}

func main() {
	// configuration := config.GetConfig(config.GetDefaultConfigFilePath())
	configuration, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	/*
		if len(os.Args) < 2 {
			helpCommand(configuration)
			os.Exit(1)
		}
	*/

	if len(os.Args) >= 2 {
		// COMMAND LINE ARGUMENTS MODE

		var (
			args           argsStruct
			searchActive   bool
			helpActive     bool
			versionActive  bool
			infoActive     bool
			setupActive    bool
			downloadActive bool
			updateActive   bool
			bulletinActive bool
			verboseActive  bool
			showActive     bool
			compareActive  bool
			plotActive     bool
			randomActive   bool

			forceDownload bool
		)

		for i := 1; i < len(os.Args); i++ {
			if strings.EqualFold(os.Args[i], "help") || strings.EqualFold(os.Args[i], "h") {
				helpActive = true
			} else if strings.EqualFold(os.Args[i], "version") || strings.EqualFold(os.Args[i], "v") {
				versionActive = true
			} else if strings.EqualFold(os.Args[i], "info") || strings.EqualFold(os.Args[i], "i") {
				infoActive = true

				setupActive = false
				searchActive = false
				showActive = false
				compareActive = false
				plotActive = false
			} else if strings.EqualFold(os.Args[i], "setup") || strings.EqualFold(os.Args[i], "e") {
				setupActive = true

				searchActive = false
				infoActive = false
				showActive = false
				compareActive = false
				plotActive = false
			} else if strings.EqualFold(os.Args[i], "download") || strings.EqualFold(os.Args[i], "d") {
				downloadActive = true

				searchActive = false
				infoActive = false
				showActive = false
				compareActive = false
				plotActive = false
				if len(os.Args) > i+1 && os.Args[i+1] == "force" {
					forceDownload = true
				}

			} else if strings.EqualFold(os.Args[i], "update") || strings.EqualFold(os.Args[i], "u") {
				updateActive = true

				searchActive = false
				infoActive = false
				showActive = false
				compareActive = false
				plotActive = false
			} else if strings.EqualFold(os.Args[i], "bulletin") || strings.EqualFold(os.Args[i], "b") {
				bulletinActive = true

				searchActive = false
				infoActive = false
				showActive = false
				compareActive = false
				plotActive = false
			} else if strings.EqualFold(os.Args[i], "setup") || strings.EqualFold(os.Args[i], "e") {
				verboseActive = true

				searchActive = false
				infoActive = false
				showActive = false
				compareActive = false
				plotActive = false
			} else if strings.EqualFold(os.Args[i], "search") || strings.EqualFold(os.Args[i], "s") {
				// toggle off active flags except for searchActive
				searchActive = true

				infoActive = false
				showActive = false
				compareActive = false
				plotActive = false

				args.search = append(args.search, searchArgs{})
			} else if strings.EqualFold(os.Args[i], "show") || strings.EqualFold(os.Args[i], "w") {
				// toggle off active flags except for showActive
				showActive = true
				searchActive = false
				infoActive = false
				compareActive = false
				plotActive = false
			} else if strings.EqualFold(os.Args[i], "compare") || strings.EqualFold(os.Args[i], "c") {
				// toggle off active flags except for compareActive
				compareActive = true
				searchActive = false
				infoActive = false
				showActive = false
				plotActive = false
			} else if strings.EqualFold(os.Args[i], "plot") || strings.EqualFold(os.Args[i], "p") {
				// toggle off active flags except for plotActive
				plotActive = true
				searchActive = false
				infoActive = false
				showActive = false
				compareActive = false
			} else if strings.EqualFold(os.Args[i], "force") || strings.EqualFold(os.Args[i], "f") {
				if os.Args[i-1] != "download" {
					fmt.Printf("Option force must come after a download command.")
					os.Exit(1)
				}
			} else if strings.EqualFold(os.Args[i], "random") || strings.EqualFold(os.Args[i], "r") {
				randomActive = true
			} else {
				if searchActive {
					// then add argument to the last element of args.search
					args.search[len(args.search)-1].terms = append(args.search[len(args.search)-1].terms, os.Args[i])
				} else if infoActive {
					args.info.active = true
					if os.Args[i] == "%" { // if % follows a plot command, infoCommand will be called upon the result from searches
						args.searchToInfo = true
					} else {
						args.info.codes = append(args.info.codes, os.Args[i])
					}
				} else if showActive {
					args.show.active = true
					if os.Args[i] == "%" { // if % follows a show command, infoCommand will be called upon the result from searches
						args.searchToShow = true
					} else {
						args.show.codes = append(args.show.codes, os.Args[i])
					}
				} else if compareActive {
					args.compare.active = true
					if os.Args[i] == "%" { // if % follows a compare command, infoCommand will be called upon the result from searches
						args.searchToCompare = true
					} else {
						args.compare.codes = append(args.compare.codes, os.Args[i])
					}
				} else if plotActive {
					args.plot.active = true
					if os.Args[i] == "%" {
						args.searchToPlot = true
					} else if os.Args[i] == "separate" || os.Args[i] == "sep" {
						args.plot.separate = true
					} else if os.Args[i] == "save" {
						args.plot.save = true
					} else {
						args.plot.codes = append(args.plot.codes, os.Args[i])
					}
				} else {
					fmt.Printf("Unrecognized argument %s\n", os.Args[i])
					os.Exit(1)
				}
			}
		}

		if verboseActive {
			fmt.Println("Verbose active")
		}
		if helpActive {
			helpCommand(configuration)
		}
		if versionActive {
			versionCommand(configuration)
		}
		if setupActive {
			setupCommand(configuration)
		}
		if downloadActive {
			downloadCommand(configuration, forceDownload)
		}
		if updateActive {
			updateCommand(configuration)
		}

		if bulletinActive {
			bulletinCommand(configuration)
		}

		if randomActive {
			randomCommand(configuration)
		}

		err = searchCommand(configuration, &args, nil)
		if err != nil {
			fmt.Printf("main: %s\n", err.Error())
		}

		infoCommand(configuration, &args)
		showCommand(configuration, &args)
		compareCommand(configuration, &args)
		plotCommand(configuration, &args)

	} else {
		// PROMPT MODE

		var quitFlag bool
		// var resultStack = make(map[string]string)
		var resultsStack = map[string]string{}

		fmt.Printf("%s %s\n", version.CmdName, version.CmdVersion)
		for {
			/*
				if resultsStack != nil {
					fmt.Printf("DUMP of resultsStack at the start of prompt loop\n")
					for codigo, titulo := range resultsStack {
						fmt.Printf("%s: %s\n", codigo, titulo)
					}
				}
			*/

			var args argsStruct
			switch commands := strings.Split(prompt.Input("> ", promptCompleter), " "); commands[0] {
			case "quit":
				quitFlag = true
			case "help":
				helpCommand(configuration)
			case "version":
				versionCommand(configuration)
			case "setup":
				setupCommand(configuration)
			case "download":
				var forceFlag bool

				if len(commands) > 1 && commands[1] == "force" {
					forceFlag = true
				}
				downloadCommand(configuration, forceFlag)
			case "update":
				updateCommand(configuration)
			case "bulletin":
				bulletinCommand(configuration)
			case "info":
				args.info.active = true
				args.info.codes = commands[1:]

				infoCommand(configuration, &args)
			case "search":
				args.search = append(args.search, searchArgs{})

				for _, command := range commands[1:] { // : len(commands)-1] {
					args.search[0].terms = append(args.search[0].terms, command)
				}

				// fmt.Printf("Search command: %v\n", args.search[0].terms)

				err = searchCommand(configuration, &args, resultsStack)
				if err != nil {
					fmt.Printf("main: %s\n", err.Error())
				}

			case "show":
				args.show.active = true

				for _, command := range commands[1:] { // : len(commands)-1] {
					if command == "%" {
						for k, _ := range resultsStack {
							args.show.codes = append(args.show.codes, k)
						}
					} else {
						args.show.codes = append(args.show.codes, command)
					}
				}

				fmt.Printf("Show command: %v\n", args.show.codes)

				showCommand(configuration, &args)
			case "compare":
				fmt.Printf("Compare command: %s\n", commands[0])
				args.compare.active = true
				args.compare.codes = commands[1:]
				compareCommand(configuration, &args)
			case "plot":
				args.plot.active = true

				for _, command := range commands[1:] { // : len(commands)-1] {
					if command == "sep" {
						args.plot.separate = true
					} else if command == "%" {
						for k, _ := range resultsStack {
							args.plot.codes = append(args.plot.codes, k)
						}
					} else {
						args.plot.codes = append(args.plot.codes, command)
					}
				}

				fmt.Printf("Plot command: %v\n", args.plot.codes)

				plotCommand(configuration, &args)
			case "random":
				fmt.Printf("Random command: %s\n", commands[0])
				randomCommand(configuration)
			}

			if quitFlag {
				break
			}
		}
	}
}
