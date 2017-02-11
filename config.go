package main

// Read in configuration files

import (
	// "encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/knoxcarey/bob/idp"
	"github.com/knoxcarey/bob/beacon"
)


// BeaconConfig holds the contents of JSON configuration file
type BeaconConfig struct {
	Beacons     []beacon.Beacon     `json:"beacons,omitempty`     // List of beacons
	IDProviders []idp.IDProvider    `json:"idProviders,omitempty` // Identity providers
}


var (
	// Configuration
	Config BeaconConfig                         

	// Location of the configuration directory
	configDir string

	// Port at which to operate service
	port int

	// Timeout for beacon queries, in seconds
	timeout int
)

const (
	defaultConfigDir  = "./config"              // Default location of config file
	defaultPort       = 8080                    // Default port for server
	defaultTimeout    = 20                      // Default timeout for queries, in seconds
)


func readBeaconConfigs() {
	beaconConfigDir := configDir + "/beacon/"
	beaconFiles, err := ioutil.ReadDir(beaconConfigDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range beaconFiles {
		beacon.AddBeaconFromConfig(beaconConfigDir + file.Name())
	}
}


func readIDPConfigs() {
}


// readConfig reads in configuration file with the given file name
func readConfig() {
	readBeaconConfigs()
	readIDPConfigs()
}


// Initialize identity provider structures
func initializeIDProviders() {
	for i, _ := range Config.IDProviders {
		idp := &(Config.IDProviders[i])
		if idp.ClientIDEnv != "" {
			idp.ClientID = os.Getenv(idp.ClientIDEnv)
		}
		if idp.ClientSecretEnv != "" {
			idp.ClientSecret = os.Getenv(idp.ClientSecretEnv)
		}
		fmt.Println(idp.ClientID)
		fmt.Println(idp.ClientSecret)
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
	readConfig()
	// beacon.InitializeBeacons(Config.Beacons)
	initializeIDProviders()
}
