/***************************************************************************
 Copyright 2017 William Knox Carey

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 ***************************************************************************/


package main

// Read in configuration files

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"github.com/knoxcarey/bob/idp"
	"github.com/knoxcarey/bob/beacon"
)


var (
	configDir string                  // Location of the configuration directory
	port int                          // Port at which to operate service
	timeout int                       // Timeout for beacon queries, in seconds
	host string                       // Host for this service
)

var (
	defaultConfigDir  = ""            // Default location of config file
	defaultPort       = 8080          // Default port for server
	defaultTimeout    = 20            // Default timeout for queries, in seconds
	defaultHost       = "127.0.0.1"   // Default host is localhost
)


// Read in configuration and apply defaults
func init() {	

	// Initialize default config directory
	_, dir, _, _ := runtime.Caller(0)
	defaultConfigDir = filepath.Join(path.Dir(dir), "config")

	// Read command line
	parseSwitches()

	// Create symbolic links for images
	linkAssets()

	// read in configuration files
	readConfigs("beacon", func (file string) {beacon.AddBeaconFromConfig(file)})
	readConfigs("idp", func (file string) {idp.AddIDPFromConfig(file)})
}


// Parse command-line switches; set defaults if not present
func parseSwitches() {
	flag.StringVar(&configDir, "config", defaultConfigDir, "Configuration directory")
	flag.StringVar(&host, "host", defaultHost, "Host name")
	flag.IntVar(&port, "port", defaultPort, "Port on which to run server")
	flag.IntVar(&timeout, "timeout", defaultTimeout, "Timeout for beacon queries, in seconds")
	flag.Parse()
}


// Make symbolic links to image assets
func linkAssets() {
	srcDir := configDir + "/img/"
	dstDir := "static/img/"
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Fatal(err)
	}
	
	for _, file := range files {
		if _, err := os.Stat(dstDir + file.Name()); os.IsNotExist(err) {
			os.Symlink(srcDir + file.Name(), dstDir + file.Name())
		}
	}
}


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

