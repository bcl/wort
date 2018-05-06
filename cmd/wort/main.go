/*wort - http API server for temperature sensor readings
  Copyright (C) 2018 Brian C. Lane <bcl@brianlane.com>

  This program is free software; you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation; either version 2 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License along
  with this program; if not, write to the Free Software Foundation, Inc.,
  51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/
package main

import (
	"flag"
	"log"
	"wort/internal/api"
	"wort/internal/db"
)

/* commandline flags */
type cmdlineArgs struct {
	DatabaseFile string // Path and filename of Bolt database for readings
	ListenIP     string // IP address to listen to
	ListenPort   int    // Port to listen to
}

/* commandline defaults */
var cfg = cmdlineArgs{
	DatabaseFile: "temperatures.db",
	ListenIP:     "0.0.0.0",
	ListenPort:   3834,
}

/* parseArgs handles parsing the cmdline args and setting values in the global cfg struct */
func parseArgs() {
	flag.StringVar(&cfg.DatabaseFile, "db", cfg.DatabaseFile, "Path and filename of database (temperatures.db)")
	flag.StringVar(&cfg.ListenIP, "ip", cfg.ListenIP, "IP Address to Listen to (0.0.0.0)")
	flag.IntVar(&cfg.ListenPort, "port", cfg.ListenPort, "Port to listen to (3834)")

	flag.Parse()
}

/* main processes cmdline arguments and starts the API server */
func main() {
	parseArgs()

	boltDb, err := db.Init(&cfg.DatabaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer boltDb.Close()

	api.Server(boltDb, &cfg.ListenIP, cfg.ListenPort)
}
