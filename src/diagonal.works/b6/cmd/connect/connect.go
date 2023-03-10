package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/region"
	"diagonal.works/b6/search"
	"diagonal.works/b6/transit"

	"github.com/golang/geo/s1"
)

func connectFeatures(features b6.Features, network transit.PathIDSet, threshold s1.Angle, w b6.World, s transit.ConnectionStrategy, cores int) {
	c := make(chan b6.Feature, cores)
	var wg sync.WaitGroup
	f := func() {
		for feature := range c {
			transit.ConnectFeature(feature, network, threshold, w, s)
		}
		wg.Done()
	}
	wg.Add(cores)
	for i := 0; i < cores; i++ {
		go f()
	}
	n := 0
	for features.Next() {
		n++
		c <- features.Feature()
	}
	close(c)
	wg.Wait()
	log.Printf("  connected %d features", n)
}

func main() {
	addr := flag.String("addr", "", "Address to listen on for status over HTTP")
	base := flag.String("base", "", "World to make connections to")
	input := flag.String("input", "", "World to make connections from")
	output := flag.String("output", "", "Output connected world")
	modifyPaths := flag.Bool("modify-paths", true, "Add new connection points to existing paths")
	networkThreshold := flag.Float64("network-threshold", 500.0, "Distance to travel before a street is considered connected")
	connectionThreshold := flag.Float64("connection-threshold", 100.0, "Distance away from entrances within which highways are considered")
	clusterThreshold := flag.Float64("cluster-threshold", 4.0, "Distance below which close connection points are merged")
	cores := flag.Int("cores", runtime.NumCPU(), "Number of cores available")
	flag.Parse()

	if *addr != "" {
		go func() {
			log.Println(http.ListenAndServe(*addr, nil))
		}()
		log.Printf("Listening on %s", *addr)
	}

	if *base == "" && *input != "" {
		*base = *input
	} else if *base != "" && *input == "" {
		*input = *base
	} else if *base == "" && *input == "" {
		log.Fatal("Must specific --base or --input")
	}

	b, err := region.ReadWorld(*base, *cores)
	if err != nil {
		log.Fatal(err)
	}

	var i b6.World
	if *input == *base {
		i = b
	} else {
		i, err = region.ReadWorld(*input, *cores)
		if err != nil {
			log.Fatal(err)
		}
	}

	highways := b6.FindPaths(search.TokenPrefix{Prefix: "highway"}, b)
	weights := transit.SimpleHighwayWeights{}
	log.Printf("Build street network")
	network := transit.BuildStreetNetwork(highways, b6.MetersToAngle(*networkThreshold), weights, nil, b)
	log.Printf("  %d paths", len(network))
	features := i.FindFeatures(search.Union{search.TokenPrefix{Prefix: "building="}, search.TokenPrefix{Prefix: "amenity="}, search.All{Token: "landuse=vacant"}})

	var strategy transit.ConnectionStrategy
	connections := transit.NewConnections()
	if *modifyPaths {
		strategy = transit.InsertNewPointsIntoPaths{
			Connections:      connections,
			World:            b,
			ClusterThreshold: b6.MetersToAngle(*clusterThreshold),
		}

	} else {
		strategy = transit.UseExisitingPoints{
			Connections: connections,
		}
	}
	log.Printf("Connect features")
	connectFeatures(features, network, b6.MetersToAngle(*connectionThreshold), b, strategy, *cores)
	log.Printf("Cluster")
	strategy.Finish()
	log.Printf(connections.String())
	log.Printf("Output")
	config := region.Config{
		OutputFilename:       *output,
		Cores:                *cores,
		WorkDirectory:        "",
		PointsWorkOutputType: region.OutputTypeMemory,
	}
	if i == b {
		if region.BuildRegionFromPBF(strategy.Output(), &config); err != nil {
			log.Fatal(err)
		}
	} else {
		overlay := ingest.NewOverlayWorld(i, b)
		if region.BuildOverlayRegion(strategy.Output(), &config, overlay); err != nil {
			log.Fatal(err)
		}
	}
}
