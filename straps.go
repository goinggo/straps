// Copyright 2013 Ardan Studios. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package straps implements a simple way to store application settings and set environment
variable in an XML document.

Configuration

Straps expects a file name straps.xml to be located in the default working directory or
in the GOPATH plus a provided relative path location.

	<straps>
	  <env name="dev">
	    <strap key="CompanyName" value="NEWCO-DEV"/>
	    <strap key="UseEmail" value="true"/>
	  </env>
	  <env name="prod">
	    <strap key="CompanyName" value="NEWCO"/>
	    <strap key="UseEmail" value="true"/>
	  </env>
	</straps>

Example Use Of Straps

	package main

	import (
		"github.com/goinggo/straps"
		"errors"
		"fmt"
	)

	func _CatchPanic(err *error, functionName string) {

		if r := recover(); r != nil {

			fmt.Printf("%s : PANIC Defered : %v", functionName, r)

			if err != nil {
				*err = fmt.Errorf("%v", r)
			}
		} else if err != nil && *err != nil {

			fmt.Printf("%s : ERROR : %v", functionName, *err)
		}
	}

	func straps_main() {

		straps.Load("PROGRAM_ENV_VARIABLE")

		fmt.Printf("Value: %s", straps.Strap("CompanyName"))
	}
*/
package straps

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	st straps // A reference to the singleton.
)

type (
	// xmlStrap maps each strap node in the straps.xml document.
	xmlStrap struct {
		XMLName xml.Name `xml:"strap"`
		Key     string   `xml:"key,attr"`
		Value   string   `xml:"value,attr"`
	}

	// xmlEnv maps each set of specified environment straps.
	xmlEnv struct {
		XMLName xml.Name   `xml:"env"`
		Name    string     `xml:"name,attr"`
		Straps  []xmlStrap `xml:"strap"`
	}

	// xmlStraps maps all the straps.
	xmlStraps struct {
		XMLName      xml.Name `xml:"straps"`
		Environments []xmlEnv `xml:"env"`
	}

	// straps maintains a cache of items found in the straps.xml file.
	straps struct {
		strapMap        map[string]string // The map of strap key value pairs.
		environmentName string            // The environment to use.
	}
)

// MustLoad retrieves application configuration.
func MustLoad(environmentVariable string, relativeGoPathLocation string) {
	// Find the location of the straps.xml file.
	strapsFilePath, err := filepath.Abs("straps.xml")

	// Open the straps.xml file.
	file, err := os.Open(strapsFilePath)
	if err != nil {
		// Check the gopath.
		goPath := os.Getenv("GOPATH")

		// Panic
		if goPath == "" {
			panic("Unable To Locate Straps File")
		}

		// Remove the ending / if it exists.
		goPath = strings.TrimRight(goPath, "/")

		// Open the straps.xml file.
		file, err = os.Open(fmt.Sprintf("%s/src/%s/straps.xml", goPath, relativeGoPathLocation))
		if err != nil {
			panic(err.Error())
		}
	}

	defer file.Close()

	// Read the straps file.
	xmlStraps := mustReadStraps(file, environmentVariable)

	// Create a straps object.
	st = straps{
		strapMap: make(map[string]string),
	}

	// Store the key/value pairs for each strap
	for _, strap := range xmlStraps {
		st.strapMap[strap.Key] = strap.Value
	}
}

// Exists returns true if the key exists else false.
func Exists(key string) (found bool) {
	_, found = st.strapMap[key]
	return found
}

// Strap returns the value for the specified key
//  key: The key lookup string defined in the straps file
func Strap(key string) string {
	return st.strapMap[key]
}

// StrapRegexp returns a slice of matching keys.
func StrapRegexp(keyRegexp string) (matches []string) {
	matches = make([]string, 0)

	find := regexp.MustCompile(keyRegexp)

	for key, value := range st.strapMap {
		if find.MatchString(key) == true {
			matches = append(matches, value)
		}
	}

	return matches
}

// StrapBool returns the value for the specified key as a bool.
func StrapBool(key string) bool {
	strap := Strap(key)

	boolean, err := strconv.ParseBool(strap)
	if err != nil {
		return false
	}

	return boolean
}

// StrapInt returns the value for the specified key as an int.
func StrapInt(key string) int {
	strap := Strap(key)

	integer, err := strconv.Atoi(strap)
	if err != nil {
		return -1
	}

	return integer
}

//** PRIVATE FUNCTIONS

// mustReadStraps reads the straps.xml file and decodes the XML.
func mustReadStraps(reader io.Reader, environmentVariable string) []xmlStrap {
	var xmlStraps xmlStraps
	if err := xml.NewDecoder(reader).Decode(&xmlStraps); err != nil {
		panic("Unable To Read Staps File")
	}

	// Identify the enviornment to use.
	var environment string
	if len(environmentVariable) > 0 {
		environment = os.Getenv(environmentVariable)
	}

	// If the environment variable does not exist, panic.
	if len(environment) == 0 {
		panic(fmt.Errorf("Environment Variable [%s] Does Not Exist", environmentVariable))
	}

	// Find the environment to use.
	var useEnv *xmlEnv
	for _, env := range xmlStraps.Environments {
		if env.Name == environment {
			useEnv = &env
			break
		}
	}

	if useEnv == nil {
		panic(fmt.Errorf("No Environment With Name %s Found", environment))
	}

	return useEnv.Straps
}
