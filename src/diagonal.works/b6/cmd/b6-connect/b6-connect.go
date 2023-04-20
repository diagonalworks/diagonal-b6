package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/graph"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"

	"github.com/golang/geo/s1"
)

func connectFeatures(features b6.Features, network graph.PathIDSet, threshold s1.Angle, w b6.World, s graph.ConnectionStrategy, cores int) {
	c := make(chan b6.Feature, cores)
	var wg sync.WaitGroup
	f := func() {
		for feature := range c {
			graph.ConnectFeature(feature, network, threshold, w, s)
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
	connect := flag.String("connect", "[#building | #amenity | #landuse=vacant]", "Feature types to connect")
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

	var query b6.Query
	expression, err := api.ParseExpression(*connect)
	if err != nil {
		log.Fatal(err)
	}
	err = api.EvaluateAndFill(expression, b6.EmptyWorld{}, functions.Functions(), functions.FunctionConvertors(), &query)
	if err != nil {
		log.Fatal(err)
	}

	b, err := compact.ReadWorld(*base, *cores)
	if err != nil {
		log.Fatal(err)
	}

	var i b6.World
	if *input == *base {
		i = b
	} else {
		i, err = compact.ReadWorld(*input, *cores)
		if err != nil {
			log.Fatal(err)
		}
	}

	highways := b6.FindPaths(b6.Keyed{"#highway"}, b)
	weights := graph.SimpleHighwayWeights{}
	log.Printf("Build street network")
	network := graph.BuildStreetNetwork(highways, b6.MetersToAngle(*networkThreshold), weights, nil, b)
	log.Printf("  %d paths", len(network))
	features := i.FindFeatures(query)

	var strategy graph.ConnectionStrategy
	connections := graph.NewConnections()
	if *modifyPaths {
		strategy = graph.InsertNewPointsIntoPaths{
			Connections:      connections,
			World:            b,
			ClusterThreshold: b6.MetersToAngle(*clusterThreshold),
		}

	} else {
		strategy = graph.UseExisitingPoints{
			Connections: connections,
		}
	}
	log.Printf("Connect features")
	connectFeatures(features, network, b6.MetersToAngle(*connectionThreshold), b, strategy, *cores)
	log.Printf("Cluster")
	strategy.Finish()
	log.Printf(connections.String())
	log.Printf("Output")
	options := compact.Options{
		OutputFilename:       *output,
		Cores:                *cores,
		WorkDirectory:        "",
		PointsWorkOutputType: compact.OutputTypeMemory,
	}
	if i == b {
		if compact.Build(strategy.Output(), &options); err != nil {
			log.Fatal(err)
		}
	} else {
		overlay := ingest.NewOverlayWorld(i, b)
		if compact.BuildOverlay(strategy.Output(), &options, overlay); err != nil {
			log.Fatal(err)
		}
	}
}
