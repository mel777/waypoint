package main

import (
	. "math"
	"fmt"
)

type GreatCircle struct {
	Loc1	Location
	Loc2	Location
	Ang12	float64 // rad
}

type Flyby struct {
	GreatCircle
	B		float64 // rad
	C		float64 // rad
	Ang23	float64 // rad
	Ang13	float64 // rad
	Nearest	float64 // rad
	Heading	float64 // deg from north (-ve west, +ve east)
}

type FlybyFilter interface {
	NearestApproach(loc1, loc2 Location) (bool, bool, *Flyby)
}

type FlybyPoint struct {
	nearestApproach func(loc1, loc2 Location) (bool, bool, *Flyby)
}

func (fp FlybyPoint) NearestApproach(loc1, loc2 Location) (bool, bool, *Flyby) {
	return fp.nearestApproach(loc1, loc2)
}

// Balance partitions for efficient parallel computation
func (locs Locations) FindPairsPassingWithinRadius(filters []FlybyFilter) []Flyby {
	nproc := len(filters)
	var j1, j2 int
	nf64 := float64(len(locs))
	nprocf64 := float64(nproc)
	if len(locs)/nproc < 100 {
		nproc = 1
		j1 = 0
		j2 = len(locs)
	} else {
		j1 = 0
		j2 = int(nf64*Sqrt(1.0/nprocf64))
	}
	ch := make(chan []Flyby)
	for i := 0; i < nproc; i++ {
		if i == nproc-1 {
			j2 = len(locs)-1
		}
		Println("Processing locations %d:%d", j1, j2)
		// Each process must get its own copy of the filter function, otherwise
		// they would corrupt each other.  goroutines use shared memory.
		go locs.findPairsPassingWithinRadius(j1, j2, filters[i], i, ch)
		j1 = j2
		j2 = int(nf64*Sqrt((float64(i)+2.0)/nprocf64))
	}

	results := []Flyby{}
	for i := 0; i < nproc; i ++ {
		result := <-ch
		for _, gc := range result {
			results = append(results, gc)
		}
	}
	return results
}

// We try to eliminate pairs from calculation (counted as "avoided").
func MakeNearestApproachFilter(loc3 Location, amax, bmax, cmax, dmax float64, dir []float64) func(loc1, loc2 Location) (bool, bool, *Flyby) {
	if len(dir) != 2 {
		Fatal("Latitude range for great circle normal must have " +
			"two numbers, but %d given", len(dir))
	}
	dir0 := DegToRad(dir[0])
	dir1 := DegToRad(dir[1])
	amax = DistToRad(amax)
	bmax = DistToRad(bmax)
	cmax = DistToRad(cmax)
	dmax = DistToRad(dmax)
	avoid := false
	var tmp Location
	var a, b, c, d, B, C, e float64
	var v0, v1, n, t Vector
	var northdev float64
	return func(loc1, loc2 Location) (bool, bool, *Flyby) {
		avoid = false
		if loc3.Lat < 0 && loc1.Lat > 0 && loc2.Lat > 0 {
			// loc1, loc2 both other side of equator
			avoid = true
		}
		if loc3.Lat > 0 && loc1.Lat < 0 && loc2.Lat < 0 {
			// loc1, loc2 both other side of equator
			avoid = true
		}
		if !avoid {
			// Order loc1, loc2, always go north to south
			if loc1.Lat > loc2.Lat {
				tmp = loc2
				loc2 = loc1
				loc1 = tmp
			}
			a, b, c = UnitSphericalTriangleSides(loc1, loc2, loc3)
			//
			//                   3            1 start, 2 end, 3 fixed, 4 nearest
			//                  _.__   c      a -> d great circles segments
			//          b   _--`  \ ```-._    A, B, C spherical triangle angles
			//          _--`     d \ __-*` 2  d meets 4 perpendicularly
			//       _-`       __--`4   (B)   ab form C, ac form B
			//     .` C  __--``      
			//    /__--``        a
			//  1                   
			if a <= amax && b <= bmax && c <= cmax {
				// Cosine rule
				C = Acos((Cos(c) - Cos(a)*Cos(b)) / (Sin(a)*Sin(b)))
				if RadToDeg(C) < 90 {
					d = Asin(Sin(b)*Sin(C)) // Sine rule
					if d <= dmax {
						B = Acos((Cos(b) - Cos(c)*Cos(a)) / (Sin(c)*Sin(a)))
						if RadToDeg(B) < 90 {
							// Check compass heading.
							// e is angular distance from start of path on gc to
							// nearest point on gc to loc3.
							e = Atan(Cos(C)*Tan(b)) // One of Napier's rules
							// Rotate vector to start point 1, around gc normal
							// vector, by angle e, to get vector to nearest
							// point (4) on gc to loc3.
							v0 = loc1.ToCartesianVector() // start
                            n = GreatCircleNormal(loc1, loc2).Norm()
							v1 = v0.RotateAround(n, e) // nearest gc pt to loc3
                            t = v1.Cross(n).Neg() // tangent to gc at this pt
							northdev, _ = t.Norm().UnitToUnitSpherical()
							// Deviation from north in rad,
							// -ve -> west, +ve -> east
							northdev = Abs(northdev)*Sign(n.Z)
							if northdev >= dir0 && northdev <= dir1 {
								fb := &Flyby{
									GreatCircle: GreatCircle{
										Loc1:	loc1,
										Loc2:	loc2,
										Ang12:	a,
									},
									B:			B,
									C:			C,
									Ang23:		c,
									Ang13:		b,
									Nearest:	d,
									Heading:	northdev,
								}
								return true, false, fb
							}
						}
					}
				}
			}
			return false, false, nil
		}
		return false, true, nil
	}
}

func (locs Locations) findPairsPassingWithinRadius(n1, n2 int, filter FlybyFilter, iproc int, parent chan []Flyby) {
	result := []Flyby{}
	N := maxPairs(n1, n2)
	n := 0 // report counter
	pc := 0 // report counter, percentage
	dpc := 5 // approx % complete at which to report status
	Ndiv := N/(100/dpc)
	found := 0
	tried := 0
	avoided := 0
	avoid := false
	fits := false
	var fb *Flyby
	for i := n1; i <= n2; i++ {
		for j := 0; j < i; j++ {
			fits, avoid, fb = filter.NearestApproach(locs[i], locs[j])
			if fits {
				result = append(result, *fb)
				found++
			}
			if avoid {
				avoided++
			}
			tried++
			n++
			// Report status
			if n > Ndiv {
				pc += dpc
				Println(
					"Process %2d, %3d%% complete, tried %10d " +
					"avoided %10d (%3.0f%%) found %10d",
					iproc, pc, tried, avoided,
					100*float64(avoided)/float64(tried), found)
				n = 0
			}
		}
	}
	parent <- result
}

func Roundmid(f float64) int {
	n := int(f)
	frac := f - float64(n)
	if frac < 0.5 {
		return n
	} else {
		return n+1
	}
}

// Sum of a sequence with constant difference is S = (n/2)*(a1+an) where n is
// the number of terms, a1 is first, an is last term.  j1, j2 range from 0.
func maxPairs(j1, j2 int) int {
	if j1 > j2 {
		Fatal("j2 (%d) must be greater than j1 (%d)", j2, j1)
	}
	return (j2-j1)*(1+j1+j2)/2
}

func Sign(x float64) float64 {
	if x < 0.0 {
		return -1.0
	}
	return 1.0
}

func RadToDeg(rad float64) float64 {
	return rad*180.0/PI
}

func DegToRad(deg float64) float64 {
	return deg*PI/180.0
}

func LatToPolar(lat float64) float64 {
	return PI_2 - DegToRad(lat)
}

func PolarToLat(polar float64) float64 {
	return RadToDeg(PI_2 - polar)
}

func LongToAzimuth(long float64) float64 {
	return PI + DegToRad(long)
}

// Returns km along surface.
func RadToDist(rad float64) float64 {
	return rad*EARTH_RAD
}

// Takes km, returns included arc angle in radians.
func DistToRad(dist float64) float64 {
	return dist/EARTH_RAD
}

func (loc Location) ToUnitSpherical() (float64, float64) {
	return LatToPolar(loc.Lat), LongToAzimuth(loc.Long)
}

func LatLongToRadians(loc Location) (float64, float64) {
	return DegToRad(loc.Lat), DegToRad(loc.Long)
}

type Vector struct {
	X	float64
	Y	float64
	Z	float64
}

func (v Vector) String() string {
	return fmt.Sprintf("(%.4f %.4f %.4f)", v.X, v.Y, v.Z)
}

type Matrix struct {
	R0	Vector
	R1	Vector
	R2	Vector
}

func (v Vector) Neg() Vector {
	return Vector{
		-v.X,
		-v.Y,
		-v.Z,
	}
}

func (v Vector) Mag() float64 {
	return Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vector) Norm() Vector {
	vmag := v.Mag()
	return Vector{
		v.X/vmag,
		v.Y/vmag,
		v.Z/vmag,
	}
}

func (v Vector) Dot(u Vector) float64 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z
}

// Angle between vectors in radians
func (v Vector) AngleWith(u Vector) float64 {
	return Acos(v.Norm().Dot(u.Norm()))
}

func (v Vector) Cross(u Vector) Vector {
	return Vector{
		v.Y*u.Z - u.Y*v.Z,
		u.X*v.Z - v.X*u.Z,
		v.X*u.Y - u.X*v.Y,
	}
}

func (m Matrix) Times(v Vector) Vector {
	return Vector{
		m.R0.X*v.X + m.R0.Y*v.Y + m.R0.Z*v.Z,
		m.R1.X*v.X + m.R1.Y*v.Y + m.R1.Z*v.Z,
		m.R2.X*v.X + m.R2.Y*v.Y + m.R2.Z*v.Z,
	}
}

func (v Vector) RotateAround(u Vector, a float64) Vector {
    return RotationMatrix(u, a).Times(v)
}

func RotationMatrix(u Vector, a float64) Matrix {
	return Matrix{
		R0: Vector{
			Cos(a) + u.X*u.X*(1-Cos(a)),
			u.X*u.Y*(1-Cos(a)) - u.Z*Sin(a),
			u.X*u.Z*(1-Cos(a)) + u.Y*Sin(a),
		},
		R1: Vector{
			u.Y*u.X*(1-Cos(a)) + u.Z*Sin(a),
			Cos(a) + u.Y*u.Y*(1-Cos(a)),
			u.Y*u.Z*(1-Cos(a)) - u.X*Sin(a),
		},
		R2: Vector{
			u.Z*u.Z*(1-Cos(a)) - u.Y*Sin(a),
			u.Z*u.Y*(1-Cos(a)) + u.X*Sin(a),
			Cos(a) + u.Z*u.Z*(1-Cos(a)),
		},
	}
}

func (v Vector) UnitToUnitSpherical() (polar, azim float64) {
	polar = Acos(v.Z)
	azim = Atan(v.Y/v.X)
	return
}

func UnitSphericalToCartesianVector(polar, azim float64) Vector {
	return Vector{
		Sin(polar)*Cos(azim),
		Sin(polar)*Sin(azim),
		Cos(polar),
	}
}

func (loc Location) ToCartesianVector() Vector {
	return UnitSphericalToCartesianVector(loc.ToUnitSpherical())
}

func GreatCircleNormal(loc1, loc2 Location) Vector {
	v := loc1.ToCartesianVector()
	u := loc2.ToCartesianVector()
	return v.Cross(u)
}

// On unit sphere, side lengths same as included angles.
func UnitSphericalTriangleSides(loc1, loc2, loc3 Location) (a, b, c float64) {
    a = loc1.ToCartesianVector().AngleWith(loc2.ToCartesianVector())
    b = loc1.ToCartesianVector().AngleWith(loc3.ToCartesianVector())
    c = loc2.ToCartesianVector().AngleWith(loc3.ToCartesianVector())
    return
}

