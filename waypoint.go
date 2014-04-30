/*
	Download files and parse from: 
	http://www.fallingrain.com
	Thank you for making your world publicly available!
	Data apparently good to 2010
*/
package main

import (
	"flag"
	"time"
	"os"
)

func main() {
	flag.Parse()
	// Help invoked at command line
	if cmdHelp {
		Help()
	}
	datapath := GetDataPath()
	if cmdDownload {
		Download(datapath)
	}
	var locs Locations
	nativepath := GetLocationCSVPath(datapath)
	finfo, err := os.Stat(nativepath)
	removeRedundancies := true
	if !cmdMake && !cmdRaw && (!os.IsNotExist(err) && !finfo.IsDir()) {
		Println("Reading native csv file for locations")
		locs = ReadNativeLocationsFile(nativepath)
		removeRedundancies = false
	} else {
		locs = LoadLocationData(datapath)
	}

	if cmdRaw {
		if len(cmdFind) > 0 {
			locs.PrintLabelMatchedLocations(cmdFind)
		} else if len(cmdNear) > 0 {
			locs.PrintNearbyLocations(cmdNear, cmdDelta)
		}
	}

	if removeRedundancies {
		locs = locs.RemoveRedundant(1000) // half-length of square in m
	}

	if len(cmdFind) > 0 {
		locs.PrintLabelMatchedLocations(cmdFind)
	} else if len(cmdNear) > 0 {
		locs.PrintNearbyLocations(cmdNear, cmdDelta)
	}
	if cmdMake {
		locs.WriteToNativeCSV(datapath)
	}

	if len(cmdTest) > 0 {
		Test(cmdTest)
	}

	// Make filters before we cull locations that aren't waypoints or airports
	filters := locs.MakeUserFilters()

	// Only use waypoints and airports for further calcs... 
	var locs2 Locations
	var nwaypoints, nairports int
	for _, loc := range locs {
		if loc.Type == LOCTYPE["Waypoint"].Tag {
			nwaypoints++
			locs2 = append(locs2, loc)
		} else if loc.Type == LOCTYPE["Airport"].Tag {
			nairports++
			locs2 = append(locs2, loc)
		}
	}
	locs = locs2
	Println("%d airports, %d waypoints", nairports, nwaypoints)

	chosen := locs.Choose()

	Println("%d locations chosen", len(chosen))
	Println("%d great circles from chosen", maxPairs(0,len(chosen)-1))
	t0 := time.Now()
	within := chosen.FindPairsPassingWithinRadius(filters)
	t1 := time.Now()
	Println("Found %d pairs fitting criteria in %v", len(within), t1.Sub(t0))
	WriteFlybysToFile(datapath, within)
}
