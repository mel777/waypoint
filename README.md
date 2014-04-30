Where in the World is MH370?
============================

Background
----------

On 8 March 2014, a plane scheduled to fly from Kuala Lumpur to Beijing, carrying
239 people, vanished without a trace.  Overnight, flight [MH370](https://en.wikipedia.org/wiki/Malaysia_Airlines_Flight_370) became one of the greatest mysteries of modern times.  In a world of such increasing technological sophistication, with so much information at our fingertips, millions were at a loss to understand how a 63 m, 220 ton jumbo jet could completely dissappear.  Grieving families wanted answers.  The response of authorities was confusing and at times contradictory.  Information from official and leaked channels only served to add to the confusion, and in the absence of information, the internet came alive with all manner of theories and conspiracies.

Two months later, despite the largest air and sea search ever undertaken, not a single trace of the aircraft have been found.  Some of the theories stayed true to the available evidence and official narrative, others, involving a conspiracy, were based on a selection of evidence, while others entertained complex possibilities inversely correlated with plausibility.  This work is in support of [a theory](http://www.mh370yat.blogspot.com) in the middle camp.  It is based on

* Acceptance of eyewitness reports from the tiny Maldives island of Kuda Huvadhoo, thousands of miles to the west of Kuala Lumpur.  Multiple people claimed to have seen a large, loud, low flying jumbo jet fly in a south or south easterly direction at 6:15 am local (sunrise) on the morning of 8 March.  Emails regarding this highly unsual event began reaching local media the next day, but were not reported for another 9 days.  There were [no scheduled flights](http://www.mh370yat.blogspot.com/2014/04/maldives-air-activity.html) that could explain the sighting,
* Acceptance of the military radar analysis from official and unofficial sources, indicating that MH370 was commandeered by a skilled pilot, flying by waypoints, who had turned off the transponder and some satellite services, and flew evasively in a north westerly,
* Rejection of the Inmarsat-based analysis said to indicate that the aircraft was last detected somewhere along an arc stretching deep into the southern Indian Ocean, up to western China.  Doppler analysis of ping signals was said to imply a southern route as most probably, leading to a massive search and recovery operation off the coast of Western Australia,
* Rejection of the idea that the plane landed somewhere intact, the result of a government effort to secure access to something of value, or of a terrorist plot to use the aircraft for some future purpose.  Given the magnitude of the mystery MH370 has created, we cannot rule out possibilities simply on the basis of their sheer scale.  However, while these theories are physically possible, the conceivable motives lack evidence and defy logic.  A sophisticated government has many, many covert and overt means of accessing anything MH370 was carrying, and the government most well equipped in this regard, that of the United States, is strongly incentivized to maintain the flying confidence of the public.  A terrorist group is hardly likely to be able to hide such a large aircraft for long from the eyes and ears of the largest nations, let alone ever use it.  

Waypoint
--------

The Maldives eyewitnesses, rejected by their own government within one day, and by Malaysian investigators the next, are potentially crucial to understanding what happened.  If the aircraft they saw was indeed MH370, everything changes.

So this Go code aimed to address a simple question.  If we were to assume, for arguments sake, that

1. The aircraft was being flown more or less on autpilot via navigational beacons (waypoints), and
2. The Kuda Huvadhoo sighting was MH370,

then which waypoints was the pilot flying between when sighted flying S/SE over this island of only 1,000 m diameter?  If there were only a few, or maybe even just a single pair of waypoints that were being used, narrowing down the flight path would have major implications.  It would not tell us where MH370 is, but could, in theory, help in knowing where to look.

This code allows a [large database](http://www.fallingrain.com/world/index.html) (Thank You Falling Rain, whoever you are!) of all the world's airports and aviation waypoints, to be downloaded, parsed and searched for great circle paths between locations fitting customised criteria.  Go is great for this purpose because it offers the benefits of a scripting language like python, but with fast compilation, fast execution and simple parallelism.  Rather than calculate every pair, a lot of work can be prudently avoided, but the geometry is precise and I'm satisfied most of the major bugs have been ironed out.

Install
-------

If you're new to Go, you can [install it](http://golang.org/doc/install), making sure to [set up your environment](http://golang.org/doc/code.html) and setting GOMAXPROCS to the number of cpus on your machine.  On a linux system, the following is typically included ~/.bashrc if Go is pointed to by a symlink like $HOME/opt/golang/go, and your source, package and binary user tree is in, say, $HOME/usr/go:

```
export GOROOT=$HOME/opt/golang/go
export PATH=${PATH}:$GOROOT/bin
export GOPATH=$HOME/usr/go
export PATH=${PATH}:$GOPATH/bin
export GOMAXPROCS=4
```

Clone the github repo into your user tree, 

     go get github.com/mel777/waypoint

Navigate to the user source directory src/github.com/mel777/waypoint and enter

    go build
	./waypoint -help

Usage
-----

The -help switch provides a list of options.  If you have an internet connection, use -w to download the source html files from Falling Rain (which are not included here).  To save a native csv of the data (included), then use -m.  The saved file will be automatically used next time to speed up initialization.

To search the database, use -f with a regular expression that will be tested against the name and ICAO, IATA and FAA code fields.  Normally there are a lot of redundancies, and multiple locations within a  square of arbitrary half-length (current 1,000 m) are culled.  To search before this process if undertaken, use the raw switch -r. 

To find all locations within a square of half-length (specified using -d, in metres) of the given labelled location, use -n.  Again, to use the raw data, add -r.  It's pretty easy to change this to a circle, if you wanted.  Also, you can start fooling around with the code if you're not overly confident by using the -t with a string label you add to the switch in tests.go.

Go to control.go and hack the functions to change the filtering.  This is where you control the primary comparative location:

```
label := "Name:Kudahuvadhoo"
loc3, exist, _ := locs.FindBy(label)
if !exist {
	Fatal("Could not find location using %q", label)
}
amax := 4000.0 // km
bmax := 4000.0 // km
cmax := 4000.0 // km
dmax := 0.75 // km
heading := []float64{-45.0-22.5, 22.5} // deg from north
```

Kuda Huvadhoo doesn't have an airport or nav beacon, so I just added a line to a file locations__supplementary.csv which contains additional Maldives airports that the original database didn't have (evidently in part because some of them only opened since 2010).

Every great circle path between two locations, loc1 and loc2 of length a, forms a spherical triangle with loc3.  All paths are arranged to go from south to north, for calculation purposes, and the distances between loc1 and loc3 and loc2 and loc3 are b and c respectively.  Limits are imposed on these distances.

The most important distance, d, is the great circle segment from loc3 meeting perpendicularly with the path from loc1 to loc2.  Here we limit it to 750 m.  The Maldives eyewitnesses claim the jet was flying over the island, which is roughly circular with a diameter of about 1,000 m, so in this case we allow the possibility that the flight path cross as far away as about 250 m from the beach.   

The witnesses claimed a heading of S/SE, so we allow a range of headings given as the deviation from north (0 deg) with negative to the west and positive to the east.  In this case, given the actual direction was southerly, we're cutting some slack and selecting only paths between (and including) SSW to ESE. 

The Choose() method in control.go lets you restrict the search space, if desired (for example, for testing).

Code
----

The only technicaly interesting bits (for me, anyway), are the use of the function closure in MakeNearestApproachFilter in math.go, the partition splitting in the method FindPairsPassingWithinRadius and the vector manipulations, together with some spherical trig, used to get the local heading at nearest approach.

The closure is nice because it allows modularity.  You can predefine the function and pass instances of it to each process.  But there are gotchas.  You have to be careful with the shared memory, which can lead to subtle errors.  In particular, you need to back off on the use of pointers.

In this case the computation can be broken into processes and performed in parallel.  Each of N locations must be matched with all the other locations, but the direction of the path is immaterial and so we can immediately eliminate about half the solution space.  In order to make each of n partitions of equal size (for the most efficient use of cpu resources and minimum calculation time), the resulting lower (or upper) triangle in an NxN matrix must be split appropriately, which is equivalent to dividing the line y = x for x = [0,1] into n x values, each with equal area.  You can convince yourself that the necessary sequence is x1 = sqrt(1/n), x2 = sqrt(2/n), ... 

Finally, after getting bogged down trying to bend spherical trigonometry to the purpose of calculating the local compass heading (which varies as you move around a great circle), I went back to vector fundamentals.  Napier's rule is handy for finding e, the distance (included angle) from loc1 to the nearest point on the path to loc3, which can then allow us to rotate the (unit) vector to loc1 around the normal to the path great circle (obtained from the cross product of the loc1 and loc2 vectors).  The tangent vector t is found from another cross product, and we arrange the sign to fit our defintion or a northern deviation.

I didn't go too nuts with optimisation, but did try to streamline things where it counts, in the closure function loop.  The closure allows us to declare most variables outside the loop in order to minimse garbage collection.  The loop itself uses nested if statements to leave the most intensive calculation for the best candidates.

Conclusion
----------

For the outcome, see the blog :)  But this was a worthwhile exercise to demonstrate that if MH370 really did fly over that little island, we know with reasonable probability where it was flying.  I hope this might also serve as a nice, compact, and relevant little Go programming case study showing command line invocation, web scraping and parsing, text file manipulation, some mathemtics, and parallel computation with an interface and closure.
