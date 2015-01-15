package main

//import "log"

import (
	//"encoding/json"
	"log"
	"math/rand"
	"time"
)

func main() {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	o := &Octotree{}
	o.Depth = 1
	o.IsLeaf = true
	o.SetBounds([6]float64{-1000.0, -1000.0, -1000.0, 1000.0, 1000.0, 1000.0})

	start := time.Now()
	for i := 0; i < 1000000; i += 1 {
		item := &Item{Mass: [4]float64{r.Float64()*2 - 1, r.Float64()*2 - 1, r.Float64()*2 - 1, r.Float64() * 2}}
		o.add(item)
	}
	elapsed := time.Since(start)
	log.Println("Octotree add*10000 took ", elapsed.Nanoseconds()/1000, "us.")

	start = time.Now()
	results := o.query([6]float64{-1000.0, -1000.0, -1000.0, 1000.0, 1000.0, 1000.0})
	elapsed = time.Since(start)

	log.Println("Octotree query took ", elapsed.Nanoseconds(), "ns.")
	log.Println("results from query:", len(results))

	/*b, err := json.Marshal(results)
	if err == nil {
		log.Println(string(b))
	}*/

	log.Println("octants:", octants)
	log.Println("deepest:", deepest)
	log.Println("octantQueries", octantQueries)
	log.Println("itemQueries", itemQueries)
	o.calculateMassDistribution()
	log.Println("Mass:", o.Mass[3], "-> X", o.Mass[0], "Y", o.Mass[1], "Z", o.Mass[2])
}
