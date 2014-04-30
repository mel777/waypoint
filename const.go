package main

import (
	"flag"
	"strings"
	"math"
)

// Command line args/option invocations
var cmdDownload bool
var cmdHelp bool
var cmdTest string
var cmdFind string
var cmdRaw bool
var cmdNear string
var cmdDelta float64
var cmdMake bool

const (
	DATA_DIR string = "data"
	NATION_INDEX_NAME string = "nations.dat"
	ADDITIONAL_LOCATIONS_NAME string = "locations_supplementary.csv"
	RESULT_NAME string = "result.csv"
	LOCATION_CSV_NAME string = "locations_native.csv"
	BASE_ADDRESS string = "http://www.fallingrain.com/world/"
	US_STATES string = "AL AK AZ AR CA CO CT DE DC FL GA HI ID IL IN IA KS KY LA ME MT NE NV NH NJ NM NY NC ND OH OK OR MD MA MI MN MS MO PA RI SC SD TN TX UT VT VA WA WV WI WY"
	AIRPORT_TAG string = "airports"
	WAYPOINT_TAG string = "waypoints"
	AIRPORT_FILE_SUFFIX string = "_airports.html"
	WAYPOINT_FILE_SUFFIX string = "_waypoints.html"
	EARTH_RAD float64 = 6371.00 // km
	PI float64 = math.Pi
	PI_2 float64 = PI/2.0
	PI_4 float64 = PI/4.0
)

func init() {
	// Command line args/options
	flag.BoolVar(&cmdHelp, "help", false, "display this info")
	flag.BoolVar(&cmdDownload, "w", false, "download html files from http://www.fallingrain.com")
	flag.StringVar(&cmdTest, "t", "", "perform ad-hoc test identified by name")
	flag.StringVar(&cmdFind, "f", "", "find location with given regex for name or code")
	flag.BoolVar(&cmdRaw, "r", false, "use raw data")
	flag.StringVar(&cmdNear, "n", "", "find location with given label within square of half-length given by -d in m")
	flag.Float64Var(&cmdDelta, "d", 0.0, "offset value used by other commands")
	flag.BoolVar(&cmdMake, "m", false, "make native location data file in csv format")
	// Fill out location types
	for _, typ := range LOCTYPE {
		typ.SourceSuffix = "/" + strings.ToLower(typ.Plural) + ".html"
		typ.LocalSuffix = "_" + strings.ToLower(typ.Plural) + ".html"
	}
}

type LocType struct {
	Tag				string
	Plural			string
	SourceSuffix	string
	LocalSuffix		string
}

var LOCTYPE map[string]*LocType = map[string]*LocType{
	"Airport": &LocType{Tag: "Airport", Plural: "airports"},
	"Waypoint": &LocType{Tag: "Waypoint", Plural: "waypoints"},
}
