package main

import (
	"os"
	"strconv"
	"strings"
	"io/ioutil"
	"path/filepath"
	"code.google.com/p/go.net/html"
	"net/http"
)

// Download waypoint and airport pages for nations in given file.
func Download(datapath string) {
	// Grab html files
	nationpath := filepath.Join(datapath, NATION_INDEX_NAME)
	byts, err := ioutil.ReadFile(nationpath)
	ifError(err)
	lines := strings.Split(string(byts), "\n")
	for _, typ := range LOCTYPE {
		for _, line := range lines {
			if len(line) > 0 {
				parts := strings.SplitN(line, " ", 2)
				code := parts[0]
				name := strings.Replace(parts[1], " ", "_", -1)
				states := []string{""}
				if code == "US" {
					states = strings.Split(US_STATES, " ")
				}
				for _, state := range states {
					var addr, filename string
					if len(state) == 0 {
						addr = BASE_ADDRESS + code + typ.SourceSuffix
						filename = filepath.Join(datapath, code + name + typ.LocalSuffix)
					} else {
						addr = BASE_ADDRESS + code + "/" + state + typ.SourceSuffix
						filename = filepath.Join(datapath, code + state + typ.LocalSuffix)
					}
					Println("Downloading %s", addr)
					resp, err := http.Get(addr)
					ifError(err)
					txt, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					ifError(err)
					byts := []byte(txt)
					Println("Saving %s", filename)
					err = ioutil.WriteFile(filename, byts, 0755)
					ifError(err)
				}
			}
		}
	}
	return
}

func parseLatLong(a string) float64 {
	if strings.Contains(a, "(") {
		a = a[:len(a)-3]
	}
	result, err := strconv.ParseFloat(a, 64)
	ifError(err)
	return result
}

func parseDegMinSecDir(deg, min, sec, dir string) float64 {
	df64, err := strconv.ParseFloat(deg, 64)
	ifError(err)
	mf64, err := strconv.ParseFloat(min, 64)
	ifError(err)
	sf64, err := strconv.ParseFloat(sec, 64)
	ifError(err)
	result := df64 + mf64/60.0 + sf64/3600.0
	if dir == "S" || dir == "W" {
		result = -result
	}
	return result
}

// Order: ICAOcode Desc Lat Long Control Kind
func parseWaypoints(n *html.Node, locs Locations, get bool, country string) Locations {
    if n.Type == html.ElementNode && n.Data == "table" {
		get = true
	}
    if get && n.Type == html.ElementNode && n.Data == "tr" {
		i := 0
		loc := Location{Type: LOCTYPE["Waypoint"].Tag, Country: country}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "td" {
				k := c.FirstChild
				if k != nil {
					if i == 0 && k.Type == html.ElementNode && k.Data == "a" {
						k = k.FirstChild
						if k == nil {
							Println("Missing waypoint code: %#v", loc)
						} else {
							loc.ICAOcode = strings.TrimSpace(k.Data)
						}
					} else {
						if k.Type == html.TextNode {
							switch i {
								case 2:
									if len(k.Data) == 0 {
										Println("Missing latitude: %#v", loc)
									} else {
										loc.Lat = parseLatLong(k.Data)
									}
								case 3:
									if len(k.Data) == 0 {
										Println("Missing longitude: %#v", loc)
									} else {
										loc.Long = parseLatLong(k.Data)
									}
								case 4:
									loc.Control = strings.TrimSpace(k.Data)
								case 5:
									loc.Kind = strings.TrimSpace(k.Data)
									locs = append(locs, loc)
							}
						}
					}
				}
				i++
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		locs = parseWaypoints(c, locs, get, country)
	}
	return locs
}

// Order: Kind ICAOcode FAAcode IATAcode Desc Name Lat Long
func parseAirports(n *html.Node, locs Locations, get bool, country string) Locations {
    if n.Type == html.ElementNode && n.Data == "table" {
		get = true
	}
    if get && n.Type == html.ElementNode && n.Data == "tr" {
		i := 0
		loc := Location{Type: LOCTYPE["Airport"].Tag, Country: country}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "td" {
				k := c.FirstChild
				if k != nil {
					if i == 5 && k.Type == html.ElementNode && k.Data == "a" {
						k = k.FirstChild
						if k == nil {
							Println("Missing airport name: %#v", loc)
						} else {
							loc.Name = strings.TrimSpace(k.Data)
						}
					} else {
						if k.Type == html.TextNode {
							k.Data = strings.TrimSpace(k.Data)
							switch i {
								case 0:
									loc.Kind = k.Data
								case 1:
									loc.ICAOcode = k.Data
								case 2:
									loc.FAAcode = k.Data
								case 3:
									loc.IATAcode = k.Data
								case 4:
									loc.Desc = k.Data
								case 6:
									if len(k.Data) == 0 {
										Println("Missing latitude: %#v", loc)
									} else {
										loc.Lat = parseLatLong(k.Data)
									}
								case 7:
									if len(k.Data) == 0 {
										Println("Missing longitude: %#v", loc)
									} else {
										loc.Long = parseLatLong(k.Data)
									}
									locs = append(locs, loc)
							}
						}
					}
				}
				i++
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		locs = parseAirports(c, locs, get, country)
	}
	return locs
}

func getFilesBySuffix(basedir, suffix string) ([]string, error) {
	var nscan int = 0
	var result []string
	fileWithSuffix := func(fpath string, fileInfo os.FileInfo, inerr error) error {
		stat, err := os.Stat(fpath)
		if err != nil {
			return err
		}
		if nscan > 0 && stat.IsDir() {
			return filepath.SkipDir
		}
		nscan++
		if strings.HasSuffix(fpath, suffix) {
			result = append(result, fpath)
		}
		return nil
	}
	err := filepath.Walk(basedir, fileWithSuffix)
	if err != nil {
		return nil, err
	}
	return result, nil
}
