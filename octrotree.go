package main

import (
	"math"
)

// up/down, north/south, east/west iota
const (
	DNW = iota
	DNE
	DSW
	DSE
	UNW
	UNE
	USW
	USE
)

// bounds iota
const (
	XMIN = iota
	YMIN
	ZMIN
	XMAX
	YMAX
	ZMAX
)

const G = 1.0

// MAX_DEPTH of the octotree
const MAX_DEPTH = 32

// MAX_ITEMS in one octotree level, if not at max depth.
const MAX_ITEMS = 1

var octants = 0
var biggest = 0
var deepest = 0
var itemQueries = 0
var octantQueries = 0

// Items that are added to the octotree must implement the Item interface
type Item struct {
	Mass [4]float64
}

// Octotree acts as a node to one level of octotree.
type Octotree struct {
	Octants []*Octotree
	Bounds  [6]float64
	Items   []*Item
	Depth   int
	IsLeaf  bool

	// Barnes hut calculations
	Mass [4]float64
}

// add
func (o *Octotree) add(items ...*Item) {
	if o.Depth > deepest {
		deepest = o.Depth
	}
	o.Items = append(o.Items, items...)
	// If octant is leaf, and it is not at max depth, and it has
	// more items than max items, split the octotree into sub-octants.
	if o.IsLeaf && o.Depth < MAX_DEPTH && len(o.Items) > MAX_ITEMS && o.Depth < MAX_DEPTH {
		o.IsLeaf = false
		if o.Octants == nil {
			o.Octants = make([]*Octotree, 8)
		}
		for octant := 0; octant < 8; octant++ {
			o.Octants[octant] = &Octotree{}
			o.Octants[octant].SetBounds(o.subOctantBounds(octant))
			o.Octants[octant].Depth = o.Depth + 1
			o.Octants[octant].IsLeaf = true
			octants++
		}
	}

	if !o.IsLeaf {
		for _, item := range o.Items {
			o.Octants[o.subOctantIndex(item)].add(item)
		}
		o.Items = nil
	}
}

// query traverses the tree and returns all items that fit inside the query bounds
func (o *Octotree) query(bounds [6]float64) []*Item {
	var results []*Item

	octantQueries++
	if o.Items != nil {
		itemQueries++
		for _, item := range o.Items {
			if (bounds[XMIN] <= item.Mass[0] && item.Mass[0] <= bounds[XMAX]) &&
				(bounds[YMIN] <= item.Mass[0] && item.Mass[1] <= bounds[YMAX]) &&
				(bounds[ZMIN] <= item.Mass[0] && item.Mass[2] <= bounds[ZMAX]) {
				results = append(results, item)
			}
		}
		return results
	}

	for _, octant := range o.Octants {
		if ((bounds[XMIN] <= octant.Bounds[XMIN] && octant.Bounds[XMIN] <= bounds[XMAX]) || (octant.Bounds[XMIN] <= bounds[XMIN] && bounds[XMIN] >= bounds[XMAX])) &&
			((bounds[YMIN] <= octant.Bounds[YMIN] && octant.Bounds[YMIN] <= bounds[YMAX]) || (octant.Bounds[YMIN] <= bounds[YMIN] && bounds[YMIN] >= bounds[YMAX])) &&
			((bounds[ZMIN] <= octant.Bounds[ZMIN] && octant.Bounds[ZMIN] <= bounds[ZMAX]) || (octant.Bounds[ZMIN] <= bounds[ZMIN] && bounds[ZMIN] >= bounds[ZMAX])) {
			results = append(results, octant.query(bounds)...)
		}

	}
	return results
}

func (o *Octotree) calculateMassDistribution() {
	if o.IsLeaf {
		for _, item := range o.Items {
			if o.Mass[3]+item.Mass[3] > 0 {
				o.Mass[0] = (o.Mass[0]*o.Mass[3] + item.Mass[0]*item.Mass[3]) / (o.Mass[3] + item.Mass[3])
				o.Mass[1] = (o.Mass[1]*o.Mass[3] + item.Mass[1]*item.Mass[3]) / (o.Mass[3] + item.Mass[3])
				o.Mass[2] = (o.Mass[2]*o.Mass[3] + item.Mass[2]*item.Mass[3]) / (o.Mass[3] + item.Mass[3])
				o.Mass[3] += item.Mass[3]
			}
		}
	} else {
		for _, octant := range o.Octants {
			octant.calculateMassDistribution()
			if o.Mass[3]+octant.Mass[3] > 0 {
				o.Mass[0] = (o.Mass[0]*o.Mass[3] + octant.Mass[0]*octant.Mass[3]) / (o.Mass[3] + octant.Mass[3])
				o.Mass[1] = (o.Mass[1]*o.Mass[3] + octant.Mass[1]*octant.Mass[3]) / (o.Mass[3] + octant.Mass[3])
				o.Mass[2] = (o.Mass[2]*o.Mass[3] + octant.Mass[2]*octant.Mass[3]) / (o.Mass[3] + octant.Mass[3])
				o.Mass[3] += octant.Mass[3]
			}
		}
	}
}

// force returns the force between two masses [x,y,z,m float64]
func force(a, b [4]float64) float64 {
	return (G * a[3] * b[3]) / math.Pow(distance(a, b), 2)
}

// distance returns distance between two points
func distance(a, b [4]float64) float64 {
	return math.Sqrt(math.Pow(b[0]-a[0], 2) + math.Pow(b[1]-a[1], 2) + math.Pow(b[2]-a[2], 2))
}

// midpoint calculates the midpoint of two float64 variables.
func midpoint(a, b float64) float64 {
	return (a + b) / 2
}

// SetBounds sets the octotrees bounds to given
// variable: [xmin,ymin,zmin,xmax,ymax,zmax float64]
func (o *Octotree) SetBounds(bounds [6]float64) {
	for i := 0; i < 6; i++ {
		o.Bounds[i] = bounds[i]
	}
}

// subOctantIndex returns the corrent index of sub-octant index (int)
// for given item, according to it's x,y,z cordinates.
func (o *Octotree) subOctantIndex(item *Item) int {
	octant := 0
	if item.Mass[0] >= midpoint(o.Bounds[XMIN], o.Bounds[XMAX]) {
		octant += 1
	}
	if item.Mass[1] >= midpoint(o.Bounds[YMIN], o.Bounds[YMAX]) {
		octant += 2
	}
	if item.Mass[2] >= midpoint(o.Bounds[ZMIN], o.Bounds[ZMAX]) {
		octant += 4
	}
	return octant
}

// subOctantBounds returns [xmin,ymin,zmin,xmax,ymax,zmax float64]
// bounds for the sub-octant at given index.
func (o *Octotree) subOctantBounds(octant int) [6]float64 {
	var bounds [6]float64
	switch octant {
	case DNW:
		// West
		bounds[XMIN] = o.Bounds[XMIN]
		bounds[XMAX] = midpoint(o.Bounds[XMIN], o.Bounds[XMAX])
		// North
		bounds[YMIN] = o.Bounds[YMIN]
		bounds[YMAX] = midpoint(o.Bounds[YMIN], o.Bounds[YMAX])
		// Down
		bounds[ZMIN] = o.Bounds[ZMIN]
		bounds[ZMAX] = midpoint(o.Bounds[ZMIN], o.Bounds[ZMAX])

	case DNE:
		// East
		bounds[XMIN] = midpoint(o.Bounds[XMIN], o.Bounds[XMAX])
		bounds[XMAX] = o.Bounds[XMAX]
		// North
		bounds[YMIN] = o.Bounds[YMIN]
		bounds[YMAX] = midpoint(o.Bounds[YMIN], o.Bounds[YMAX])
		// Down
		bounds[ZMIN] = o.Bounds[ZMIN]
		bounds[ZMAX] = midpoint(o.Bounds[ZMIN], o.Bounds[ZMAX])

	case DSW:
		// West
		bounds[XMIN] = o.Bounds[XMIN]
		bounds[XMAX] = midpoint(o.Bounds[XMIN], o.Bounds[XMAX])
		// South
		bounds[YMIN] = midpoint(o.Bounds[YMIN], o.Bounds[YMAX])
		bounds[YMAX] = o.Bounds[YMAX]
		// Down
		bounds[ZMIN] = o.Bounds[ZMIN]
		bounds[ZMAX] = midpoint(o.Bounds[ZMIN], o.Bounds[ZMAX])

	case DSE:
		// East
		bounds[XMIN] = midpoint(o.Bounds[XMIN], o.Bounds[XMAX])
		bounds[XMAX] = o.Bounds[XMAX]
		// South
		bounds[YMIN] = midpoint(o.Bounds[YMIN], o.Bounds[YMAX])
		bounds[YMAX] = o.Bounds[YMAX]
		// Down
		bounds[ZMIN] = o.Bounds[ZMIN]
		bounds[ZMAX] = midpoint(o.Bounds[ZMIN], o.Bounds[ZMAX])

	case UNW:
		// West
		bounds[XMIN] = o.Bounds[XMIN]
		bounds[XMAX] = midpoint(o.Bounds[XMIN], o.Bounds[XMAX])
		// North
		bounds[YMIN] = o.Bounds[YMIN]
		bounds[YMAX] = midpoint(o.Bounds[YMIN], o.Bounds[YMAX])
		// Up
		bounds[ZMIN] = midpoint(o.Bounds[ZMIN], o.Bounds[ZMAX])
		bounds[ZMAX] = o.Bounds[ZMAX]

	case UNE:
		// East
		bounds[XMIN] = midpoint(o.Bounds[XMIN], o.Bounds[XMAX])
		bounds[XMAX] = o.Bounds[XMAX]
		// North
		bounds[YMIN] = o.Bounds[YMIN]
		bounds[YMAX] = midpoint(o.Bounds[YMIN], o.Bounds[YMAX])
		// Up
		bounds[ZMIN] = midpoint(o.Bounds[ZMIN], o.Bounds[ZMAX])
		bounds[ZMAX] = o.Bounds[ZMAX]

	case USW:
		// West
		bounds[XMIN] = o.Bounds[XMIN]
		bounds[XMAX] = midpoint(o.Bounds[XMIN], o.Bounds[XMAX])
		// South
		bounds[YMIN] = midpoint(o.Bounds[YMIN], o.Bounds[YMAX])
		bounds[YMAX] = o.Bounds[YMAX]
		// Up
		bounds[ZMIN] = midpoint(o.Bounds[ZMIN], o.Bounds[ZMAX])
		bounds[ZMAX] = o.Bounds[ZMAX]

	case USE:
		// East
		bounds[XMIN] = midpoint(o.Bounds[XMIN], o.Bounds[XMAX])
		bounds[XMAX] = o.Bounds[XMAX]
		// South
		bounds[YMIN] = midpoint(o.Bounds[YMIN], o.Bounds[YMAX])
		bounds[YMAX] = o.Bounds[YMAX]
		// Up
		bounds[ZMIN] = midpoint(o.Bounds[ZMIN], o.Bounds[ZMAX])
		bounds[ZMAX] = o.Bounds[ZMAX]
	}
	return bounds
}
