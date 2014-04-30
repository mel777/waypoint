package main

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"path/filepath"
	"code.google.com/p/go.net/html"
	"encoding/csv"
)

func GetDataPath() string {
    // Where am I?
	cwd, err := os.Getwd()
	ifError(err)
	return filepath.Join(cwd, DATA_DIR)
}

func GetResultPath(datapath string) string {
	return filepath.Join(datapath, RESULT_NAME)
}

func GetLocationCSVPath(datapath string) string {
	return filepath.Join(datapath, LOCATION_CSV_NAME)
}

func LoadLocationData(datapath string) Locations {
	locs := Locations([]Location{})
	// Parse waypoint html files
	filenames, err :=
		getFilesBySuffix(datapath, LOCTYPE["Waypoint"].LocalSuffix)
	ifError(err)
	Println("Parsing %d waypoint files", len(filenames))
	for _, filename := range filenames {
		//Println(" %s\n", filename)
		file, err := os.Open(filename)
		ifError(err)
		defer file.Close()
		doc, err := html.Parse(file)
		locs = parseWaypoints(
			doc, locs, false,
			filepath.Base(
				strings.TrimSuffix(filename, LOCTYPE["Waypoint"].LocalSuffix)))
	}
	// Parse airport html files
	filenames, err = getFilesBySuffix(datapath, LOCTYPE["Airport"].LocalSuffix)
	ifError(err)
	Println("Parsing %d airport files", len(filenames))
	for _, filename := range filenames {
		file, err := os.Open(filename)
		ifError(err)
		defer file.Close()
		doc, err := html.Parse(file)
		locs = parseAirports(
			doc, locs, false,
			filepath.Base(
				strings.TrimSuffix(filename, LOCTYPE["Airport"].LocalSuffix)))
	}
	// Parse additional locations in csv file
	otherpath := filepath.Join(datapath, ADDITIONAL_LOCATIONS_NAME)
	locs2 := readSupplementaryLocationsFromFile(otherpath)
	Println("Parsed %d additional locations in %s", len(locs2), otherpath)
	for _, loc := range locs2 {
		locs = append(locs, loc)
	}
	return locs
}

// CSV file, written before I discovered encoding/csv
func readSupplementaryLocationsFromFile(fpath string) Locations {
	result := Locations([]Location{})
	byts, err := ioutil.ReadFile(fpath)
	ifError(err)
	lines := strings.Split(string(byts), "\n")
	var latdeg, latmin, latsec, longdeg, longmin, longsec string
	var latdir, longdir string
	for _, line := range lines {
		if len(line) > 0 {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				continue
			}
			fields := strings.Split(line, ",")
			loc := Location{}
			for j, field := range fields {
				field = strings.Trim(field, ` \"'`)
				switch j {
					case 0: loc.Type = field
					case 1: loc.Country = field
					case 2: loc.ICAOcode = field
					case 3: loc.IATAcode = field
					case 4: loc.Name = field
					case 5: loc.Desc = field
					case 6: latdeg = field
					case 7: latmin = field
					case 8: latsec = field
					case 9: latdir = field
					case 10: longdeg = field
					case 11: longmin = field
					case 12: longsec = field
					case 13: longdir = field
				}
			}
			loc.Lat = parseDegMinSecDir(latdeg, latmin, latsec, latdir)
			loc.Long = parseDegMinSecDir(longdeg, longmin, longsec, longdir)
			result = append(result, loc)
		}
	}
    return result
}

func WriteFlybysToFile(datapath string, flybys []Flyby) {
	if len(flybys) > 0 {
		resultpath := GetResultPath(datapath)
		err := os.RemoveAll(resultpath)
		ifError(err)
		Println("Writing results to %s", resultpath)
		out, err := os.Create(resultpath)
		defer out.Close()
		ifError(err)
		var line string
		for i, fb := range flybys {
			line = fmt.Sprintf("%d,%s,%s,%.1f,%.3f,%.1f,%.1f,%.1f\n",
				i, fb.Loc1.ToSimpleCSV(), fb.Loc2.ToSimpleCSV(),
				RadToDist(fb.Ang12), RadToDist(fb.Nearest),
				RadToDeg(fb.B), RadToDeg(fb.C), RadToDeg(fb.Heading))
			out.WriteString(line)
		}
	}
    return
}

func (locs Locations) WriteToNativeCSV(datapath string) {
	path := GetLocationCSVPath(datapath)
	err := os.RemoveAll(path)
	ifError(err)
	Println("Writing location data to %s", path)
	out, err := os.Create(path)
	defer out.Close()
	ifError(err)
	for _, loc := range locs {
		out.WriteString(loc.ToCSVLine())
	}
    return
}

// Assume file exists
func ReadNativeLocationsFile(nativepath string) Locations {
	result := Locations([]Location{})
	file, err := os.Open(nativepath)
	ifError(err)
	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	ifError(err)
	var loc Location
	for i, fields := range lines {
		loc = FromCSVLine(fields, i)
		result = append(result, loc)
	}
    return result
}
