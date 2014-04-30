package main

import (
	"fmt"
	"math"
	"time"
	"os"
	"regexp"
	"strings"
	"strconv"
)

type Location struct {
	Type		string
	Lat			float64 // deg
	Long		float64 // deg
	ICAOcode	string
	Country		string
	State		string
	Region		string
	// Airport
	Kind		string
	FAAcode		string
	IATAcode	string
	Desc		string
	Name		string
	// Waypoint
	Control     string
}

func NewLocation() *Location {
	return &Location{}
}

func (loc Location) String() string {
	id := ""
	if len(loc.IATAcode) > 0 {
		id = "IATA:" + loc.IATAcode
	} else if len(loc.ICAOcode) > 0 {
		id = "ICAO:" + loc.ICAOcode
	} else {
		id = "Name:" + loc.Name
	}
	return fmt.Sprintf("%s %s %f %f", loc.Type, id, loc.Lat, loc.Long)
}

func (loc Location) ToCSVLine() string {
	return fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%.6f,%.6f\n",
		FixText(loc.Type),			// 0
		FixText(loc.Country),		// 1
		FixText(loc.State),			// 2
		FixText(loc.Region),		// 3
		FixText(loc.ICAOcode),		// 4
		FixText(loc.IATAcode),		// 5
		FixText(loc.FAAcode),		// 6
		FixText(loc.Name),			// 7
		FixText(loc.Kind),			// 8
		FixText(loc.Desc),			// 9
		FixText(loc.Control),		// 10
        loc.Lat,					// 11
		loc.Long)					// 12
}

func FixText(txt string) string {
	result := strings.Replace(txt, ",", " ", -1)
	result = strings.Replace(result, "\"", "", -1)
	return result
}

func FromCSVLine(parts []string, n int) Location {
	if len(parts) != 13 {
		Fatal("Line %d has %d, not 13 columns", n, len(parts))
	}
	latstr := parts[len(parts)-2]
	longstr := parts[len(parts)-1]
	lat, err := strconv.ParseFloat(latstr, 64)
	ifError(err)
	long, err := strconv.ParseFloat(longstr, 64)
	ifError(err)
	return Location{
		Type:		parts[0],
		Country:	parts[1],
		State:		parts[2],
		Region:		parts[3],
		ICAOcode:	parts[4],
		IATAcode:	parts[5],
		FAAcode:	parts[6],
		Name:		parts[7],
		Kind:		parts[8],
		Desc:		parts[9],
		Control:	parts[10],
		Lat:		lat,
		Long:		long,
	}
}

type Locations []Location

// Consider a location redundant if another exists within a 2dx*2dx square of
// it, dx in metres, taken over the Earth's surface.
func (locs Locations) RemoveRedundant(d float64) Locations {
	n0 := len(locs)
	Println(
		"Removing redundancies from %d raw locations, might take a while", n0)
	t0 := time.Now()
	ddeg := RadToDeg(DistToRad(d/1000.0))
	keep := []int{}
	found := false
	var i, j int
	for i = 0; i<len(locs); i++ {
		found = false
		for j = 0; j < i; j++ {
			if math.Abs(locs[i].Lat-locs[j].Lat) < ddeg &&
				math.Abs(locs[i].Long-locs[j].Long) < ddeg {
				found = true
				// Favor keeping airports over waypoints
				if locs[i].Type == LOCTYPE["Waypoint"].Tag &&
					locs[j].Type == LOCTYPE["Airport"].Tag {
					locs[i] = locs[j]
				}
				break
			}
		}
		if !found {
			keep = append(keep, i)
		}
	}
	result := Locations([]Location{})
	for _, i = range keep {
		result = append(result, locs[i])
	}
	t1 := time.Now()
	n1 := len(result)
	Println(
		"Now have %d locations, removed %d (%.1f%%) in %v",
		n1, n0-n1, 100.0*float64(n0-n1)/float64(n0), t1.Sub(t0))
	return result
}

// Use \b to delineate whole word, e.g. "\bMyWord\b", otherwise this function
// will find locations with name/codes containing "MyWord".
func (locs Locations) PrintLabelMatchedLocations(regex string) {
	Println(
		"Results matching regular expression %#q from %d locations:",
		regex, len(locs))
	match := false
	found := false
	var err error
	for _, loc := range locs {
		found = false
        match, err = regexp.MatchString(regex, loc.Name)
		ifError(err)
		found = match
        match, err = regexp.MatchString(regex, loc.ICAOcode)
		ifError(err)
		found = match || found
        match, err = regexp.MatchString(regex, loc.IATAcode)
		ifError(err)
		found = match || found
        match, err = regexp.MatchString(regex, loc.FAAcode)
		ifError(err)
		found = match || found
		if found {
            Println(" %v", loc)
		}
	}
    os.Exit(0)
}

// Use \b to delineate whole word, e.g. "\bMyWord\b", otherwise this function
// will find locations with name/codes containing "MyWord".
func (locs Locations) PrintNearbyLocations(label string, d float64) {
	loc0, exist, i0 := locs.FindBy(label)
	if !exist {
		Println("No location found with that label code")
		os.Exit(1)
	}
	Println(
		"Results within a %.1f m by %.1f m square centered on %v:",
		2*d, 2*d, loc0)
	ddeg := RadToDeg(DistToRad(d/1000.0))
	c := 0
	for i := 0; i<len(locs); i++ {
		if i != i0 {
			if math.Abs(locs[i].Lat-loc0.Lat) < ddeg &&
				math.Abs(locs[i].Long-loc0.Long) < ddeg {
				Println(" %v", locs[i])
				c++
			}
		}
	}
	Println(
		"Found %d locations, which have %d unique pairs",
		c, maxPairs(0, c-1))
    os.Exit(0)
}

func (locs Locations) FindByICAO(icaoCode string) (Location, bool, int) {
	for i, loc := range locs {
		if loc.ICAOcode == icaoCode {
			return loc, true, i
		}
	}
	return Location{}, false, 0
}

// Provide a label such as "Name:WAYPT1" or "ICAO:GAN".
func (locs Locations) FindBy(label string) (Location, bool, int) {
	if !strings.Contains(label, ":") {
		Fatal("%q does not contain a :", label)
	}
	parts := strings.Split(label, ":")
	if len(parts) > 2 {
		Fatal("%q has too many : characters", label)
	}
	var field int // speed up a little using integer
	switch strings.ToUpper(parts[0]) {
		case "NAME": field = 0
		case "ICAO": field = 1
		case "IATA": field = 2
		case "FAA": field = 3
	}
	value := parts[1]
	for i, loc := range locs {
		switch field {
			case 0: if loc.Name == value { return loc, true, i }
			case 1: if loc.ICAOcode == value { return loc, true, i }
			case 2: if loc.IATAcode == value { return loc, true, i }
			case 3: if loc.FAAcode == value { return loc, true, i }
		}
	}
	return Location{}, false, 0
}
