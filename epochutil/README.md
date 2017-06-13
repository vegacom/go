epochutil
=========

Binary epochutil formats and parses dates. This is a command line
interface tool for the Go time package.

Install
-------

go get -v github.com/vegacom/go/epochutil

Use cases
---------

Print the time now in the default timezone:

  epochutil now

Print the time now in Mexico City:

  epochutil --zone mexico_city now

Print the time four hours ago:

  epochutil --delta -4h now

Parse the epoch time in nanoseconds, print the time:

  epochutil 1388586612345678901

Parse the epoch time in seconds, and print the time:

  epochutil 1388586612

Parse the date and print the time:

  epochutil "2014-01-01"

Parse the date and print the time one day after in Mexico City:

  epochutil --delta 24h --zone mexico_city "2014-01-01 09:30:10.267 +0900 JST"
