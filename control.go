package main

func (locs Locations) MakeUserFilters() []FlybyFilter {
	// Now for some great circles
	label := "Name:Kudahuvadhoo"
	loc3, exist, _ := locs.FindBy(label)
	if !exist {
		Fatal("Could not find location using %q", label)
	}
	amax := 4000.0 // km
	bmax := 2000.0 // km
	cmax := 2000.0 // km
	dmax := 0.5 //0.75 // km
	heading := []float64{-45.0,0}//-45.0-22.5, 22.5} // deg from north

	nproc := 4
	filters := []FlybyFilter{}
	for i := 0; i < nproc; i++ {
		ff := &FlybyPoint{
			nearestApproach: MakeNearestApproachFilter(loc3, amax, bmax, cmax, dmax, heading),
		}
		filters = append(filters, ff)
	}
    return filters
}

func (locs Locations) Choose() Locations {
	//choose := []string{"ICAO:VRMV", "ICAO:GAN"}
	choose := []string{}
	var chosen Locations
	if len(choose) > 0 {
		for _, label := range choose {
			loc, exists, _ := locs.FindBy(label)
			if exists {
				chosen = append(chosen, loc)
			}
		}
	} else {
		chosen = locs //locs[1:10000]
	}
	return chosen
}
