package main

// Read in configuration files

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/knoxcarey/bob/idp"
	"github.com/knoxcarey/bob/beacon"
)


var (
	configDir string                  // Location of the configuration directory
	port int                          // Port at which to operate service
	timeout int                       // Timeout for beacon queries, in seconds
)

const (
	defaultConfigDir  = "./config"    // Default location of config file
	defaultPort       = 8080          // Default port for server
	defaultTimeout    = 20            // Default timeout for queries, in seconds
)


// Read configuration files from a subdirectory and perform action on each
func readConfigs(subdir string, action func (file string)) {
	directory := configDir + "/" + subdir + "/"
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		action(directory + file.Name())
	}
}


// Parse command-line switches; set defaults if not present
func parseSwitches() {
	flag.StringVar(&configDir, "config", defaultConfigDir, "Configuration directory")
	flag.IntVar(&port, "port", defaultPort, "Port on which to run server")
	flag.IntVar(&timeout, "timeout", defaultTimeout, "Timeout for beacon queries, in seconds")
	flag.Parse()
}



// Read in configuration and apply defaults
func init() {	
	parseSwitches()
	readConfigs("beacon", func (file string) {beacon.AddBeaconFromConfig(file)})
	readConfigs("idp", func (file string) {idp.AddIDPFromConfig(file)})
}
