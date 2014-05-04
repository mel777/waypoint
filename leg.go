package main

type Leg struct {
	Loc1		Location
	Loc2		Location
}

type Legs []Leg

func (locs Locations) ToLegs() Legs {
	legs := Legs([]Leg{})
	if len(locs) > 1 {
		for i := 0; i < len(locs)-1; i++ {
			legs = append(legs, Leg{Loc1: locs[i], Loc2: locs[i+1]})
		}
	}
	return legs
}

func (locs Locations) Path(labstr, datapath string) {
	locs2 := locs.LabelsToLocations(labstr)
	legs := locs2.ToLegs()
	if len(legs) > 0 {
		var dist float64
		for _, leg := range legs {
			dist += RadToDist(
				leg.Loc1.ToCartesianVector().AngleWith(
					leg.Loc2.ToCartesianVector()))
		}
		Println(
			"Great circle path length across %d locations is %.1f km",
			len(locs2), dist)
		locs2.WriteToSimpleCSV(datapath)
	} else {
		Println("Cannot process path for a single location")
	}
	return
}

