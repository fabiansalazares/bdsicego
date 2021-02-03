
# General
* [ ] Python bindings
* [ ] R bindings
* [ ] Currently, NaN values get converted to -99999.9 since Go standard library implementation of decode/json cannot marshal NaN codes. Find an alternative implementation that can deal with this.
* [ ] Introduce ffjson to optimize marshaling and unmarshaling json data 
* [x] Configuration option to set img viewer that will be called to show plots.
* [x] Add a short form for each command
* [x] Add case-insensivity for each command
* [x] Add prompt interface if called with no arguments. Use library [https://github.com/c-bata/go-prompt](Enlace)
* [x] Add err return value to config.GetConfig and check for an existing config file. If it doesn't exist, prompt an offer to setup a config file.
	- [x] Check for operating system and adjust default image viewer accordingly.
* [x] Write tests for all subpackages
	- [x] Write the test for fabiansalazares/bdsicego/database
	- [x] Write the test for fabiansalazares/bsdicego/download
	- [x] Write the test for fabiansalazares/bdsicego/series
	- [x] Write the test for fabiansalazares/bdsicego/decode

# Commands
* [x] Download
	- [x] Add progress bar for download
	- [x] Add progress bar for database build
* [ ] Update
	- [ ] Keep track of previously downloaded updates in a json file instead of checking extract folder names
	- [ ] Check for updates available but not downloaded and offer to perform a full download instead of an update.
	- [ ] Compare date of latest available update and full database download. Do not download if they match
* [ ] Bulletin
	- [ ] Implement bulletin download in download.go
	- [ ] Implement bulletin command in cmd/bdsicego
* [ ] Compare:
 	- [ ] Write command, possibly re-using code from showCommand
 	- [ ] It should show the series side-by-side, adjusting rows so that each rows the observation for all the compared series, regardless of the range of each serie and their respective frequencies.
* [x] Show
	- [ ] Paginate results while keeping coloring ('less' shell command not an option)
	- [x] Add footer including serie code and title
	- [ ] Calculate the maximum width of the table using len(serie.Title)
	- [x] Growth column should check whether the units of the series are percentage values. If so, growth should be calculated as the difference between periods, not the difference divided by the period -t
	- [x] Show historical range of the serie
* [-] Plot
	- [x] Write sketch of command with working alternative colors and ticker with dates
	- [x] joint modifier should be removed altogether. All codes given to plot should be plot jointly, but multiple plots should be allowed in one single command line call. Optionally, a separate modifier should be introduced 
	- [x] Separate modifier introduced
	- [x] Implement separate plots
	- [x] Call to image viewer should result in a fully forked process, so that bds process does not way for the viewer to finish. This is crucial in order to implement separate plots option.
	- [ ] Increase frequency of dates and tickers.
	- [ ] If series to plot from % are too many (define too many first), users should be given the option to choose which ones will be plot and which ones will not.
	- [ ] Implement modifier to save plot to specified file instead of a /tmp file 
	- [ ] Fix plot title showing up even when there is more than one serie to plot.
	- [ ] Define plotting style at plotting.yml or something similar and load before plotting.
	- [ ] Correct a bug concerning series with daily observations. They are currently plotted as a "compressed" version. It's possible to see this bug in action plotting together series 634814 and 634814q (search string PRECIO PETROLEO BRENT)
* [x] Info
	- [ ] Add nicer formatting using pretty-table
* [x] Search
	- [x] Change the system of searches so that multiple searches can be performed and added to other commands via %
 	- [x] Include feature to exclude series that match terms with a "-" or "not"
	- [x] Fix case insensitivity problem for serie codes/re-locate managing of case and diacritics to database.go:Search() away from bds.go:searchCommand()
	- [ ] Include a feature to match alternative terms
	- [ ] Sort results by code, lexicographically
* [ ] Range
	- [ ] Limit range shown by showCommand and plotted by plotCommand
* [x] Random serie command
	- [x] Currently, a BDSICEDatabase object stores the codes and their corresponding titles as a map[string]string. In order to pick a random serieCode, a slice containing the keys of the map[string]string has to be extracted every time the random command is invoked. Given that the random command is expected to be run regularly (otherwise, it would not be a command), it would be desirable to include an array containing all the serie codes in the BDSICEDatabase and in the interface definition. 
	- [ ] Pick only random series for which the most recent value refers to current or previous year.

