# bdsicego

bdsicego is a client for [BDSICE](http://serviciosede.mineco.gob.es/Indeco/BDSICE/HomeBDSICE.aspx) written in Go. [BDSICE](http://serviciosede.mineco.gob.es/Indeco/BDSICE/HomeBDSICE.aspx) stands for Base de Datos de Series de Coyuntura Económica and it is a database compiled by the Ministry of the Economy of Spain. It contains 22000+ data series relevant to the Spanish economy.

While the current website contains a basic interface to it and it is possible to download the full database and daily updates encoded in an ad-hoc format called "Xeriex" ([(see here)](http://serviciosede.mineco.gob.es/Indeco/BDSICE/formatoXeriex.aspx)), there is no currently no available API in which json data can be downloaded on-demand. bdsicego covers that gap and provides a Go library to decode Xeriex .xer files into JSON files that can be readily read into Python, R or any other data analysis libraries (along with other Go functions for basic data analysis of the database). It also provides a bdsicego executable command-line command that allows for quick exploration of the database: searching, displaying and plotting series.

bdsicego 0.1 has been tested on Linux (Debian Buster 10) and Windows 10. Code has been written with portability in mind, but it has not been yet compiled and tested for Mac platform. Let me know if you are interested on trying it and want to test it on a Mac computer.

![](search-and-plot-example.gif)

## Installing bdsicego command line util

```shell
$ git clone git@github.com:fabiansalazares/bdsicego.git
$ cd bdsicego/cmd/bdsicego/
$ go build
$ ./bdsicego
```


