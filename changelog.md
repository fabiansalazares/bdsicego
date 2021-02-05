# 06 02 2021
* Written basic README


# 28 01 2021
0.0.37
* Prompt mode: Search command now working.
* Prompt mode: Plot and Show commands now loading previous results if an argument is "%"

# 21 01 2021
0.0.36
* Info and plot command do not panic anymore when any serie code does not exist.
* Prompt interface commands info, show, plot now working

# 14 01 2021
0.0.35
* prompt interface working

# 13 01 2021
0.0.34
* download command now checks whether the database has already been downloaded or not. If a database exists, the user is prompted to confirm the download unless 'force' modifier has been passed.
* config.GetConfig() now returns an error value if config.yml file cannot be parsed.
* If no config.yml file exists in platform-dependent configuration path, prompt the user to create a default-valued config.yml file or proceed with default values.
* Adjust default viewer file depending on current platform.
* Fixes a bug related to slashes and inverted slashes when unzipping files on Windows. path/filepath should be used instead of path when joining paths on the filesystem.
* Temporary files containing the plots are now stored in a platform-dependent folder.

# 12 01 2021
0.0.33
* Configuration package now checks for an existing configuration file at platform-dependant standard location. If no configuration file is found, the user is prompted with a choice to create one with default values or just continue with default hard-coded configuration.
* Ensure that db folder within data folder exists.
 
# 10 01 2021
0.0.32
* Refactored download/DownloadFullDatabase() and download/Update(). Currently both working and clean messages.
* Added tests for database, download, series and decode subpackages.

# 24 06 2020
0.0.31
* Added start and end period of a serie to showCommand output.

# 21 06 2020

0.0.30
* Plots are now shown in viewers without waiting for them to end. Implementation of separate plotting should now be easier.
* Separate plots has now been implemented. "sep" modifier to plot commands triggers separate plotting.
* Fixed some panics when some commands were called without arguments.

0.0.29
* Multiple searches implemented and working. Example: 
	
	$ bds s avila ocupados s burgos ocupados plot %  <- plot together the results from two searches "avila ocupados" and "burgos ocupados"
* Exclude search terms by appending "-" working. Case insensitivity for search terms was managed at bds.go:searchCommand level. So all the search terms, to match and to exclude, were in capital letters when they were passed as arguments to database:Search(). When we checked serie codes against match and exclude terms, we have a capital letters term but code remains lowercase if it is originally so. Case insensitivity and diacritics should perhaps be managed at Search() level. Now, case and diacritics transformation is managed at the database:Search() level, which should also be more coherent in the future when we search in other databases with different title and code case conventions.
