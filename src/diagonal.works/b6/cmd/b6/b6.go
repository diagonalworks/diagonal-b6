package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	rpprof "runtime/pprof"
	"sync"
	"strings"
	"errors"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	b6grpc "diagonal.works/b6/grpc"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/ui"

	"google.golang.org/grpc"

	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/gcs"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
)

func main() {
	httpFlag := flag.String("http", ":8001", "Host and port on which to serve HTTP")
	grpcFlag := flag.String("grpc", ":8002", "Host and port on which to serve GRPC")
	grpcSizeFlag := flag.Int("grpc-size", 16*1024*1024, "Maximum size for GRPC messages")
	worldFlag := flag.String("world", "", "World to load")
	readOnlyFlag := flag.Bool("read-only", false, "Prevent changes to the world")
	staticFlag := flag.String("static", "src/diagonal.works/b6/cmd/b6/js/static", "Path to static content")
	jsFlag := flag.String("js", "src/diagonal.works/b6/cmd/b6/js", "Path to JS bundle")
	staticV2Flag := flag.String("static-v2", "frontend/dist", "Path to V2 static content")
	storybookFlag := flag.String("storybook", "frontend/dist-storybook", "Path to V2 static content")
	enableV2UIFlag := flag.Bool("enable-v2-ui", false, "Use the V2 UI by default")
	enableStorybookFlag := flag.Bool("enable-storybook", false, "Serve storybook on /storybook")
	enableViteFlag := flag.Bool("enable-vite", false, "Serve javascript from a development vite server")
	coresFlag := flag.Int("cores", runtime.NumCPU(), "Number of cores available")
	fileIOFlag := flag.Bool("file-io", true, "Is file IO allowed from the API?")

	additionalWorlds := make(map[b6.FeatureID]b6.World)
		flag.Func("add-world", "Additional worlds; specify like \"<feature_id> <world-arguments>\"", func(s string) error {
			featureIdStr, worldStr, found := strings.Cut(s, " ")
			if (found) {
				world, err := compact.ReadWorld(worldStr, &ingest.BuildOptions{Cores: *coresFlag})
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					return err
				}
				featureId := b6.FeatureIDFromString(featureIdStr)
				additionalWorlds[featureId] = world
				return nil
			}
			return errors.New(fmt.Sprintf("Couldn't load additional world; bad string; expected one space: %s", s))
	})

	flag.Parse()

	if *worldFlag == "" {
		fmt.Fprintln(os.Stderr, "Must specify --world")
		os.Exit(1)
	}

	base, err := compact.ReadWorld(*worldFlag, &ingest.BuildOptions{Cores: *coresFlag})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	additionalMutableWorlds := make(map[b6.FeatureID]ingest.MutableWorld)
	for featureId, world := range additionalWorlds {
		log.Printf("Adding new world at %s", featureId)
		additionalMutableWorlds[featureId] = ingest.NewMutableOverlayWorld(ingest.NewOverlayWorld(world, base))
	}

	var worlds ingest.Worlds
	if *readOnlyFlag {
		worlds = ingest.ReadOnlyWorlds{Base: base}
	} else {
		worlds = &ingest.MutableWorlds{Base: base, Mutable: additionalMutableWorlds}
	}

	apiOptions := api.Options{
		Cores:         *coresFlag,
		FileIOAllowed: *fileIOFlag,
	}

	var lock sync.RWMutex

	options := ui.Options{
		StaticPath:      *staticFlag,
		JavaScriptPath:  *jsFlag,
		StaticV2Path:    *staticV2Flag,
		StorybookPath:   *storybookFlag,
		EnableV2UI:      *enableV2UIFlag,
		EnableVite:      *enableViteFlag,
		EnableStorybook: *enableStorybookFlag,
		Worlds:          worlds,
		APIOptions:      apiOptions,
		Lock:            &lock,
	}

	handler := http.NewServeMux()
	if err := ui.RegisterWebInterface(handler, &options); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	ui.RegisterTiles(handler, &options)

	handler.HandleFunc("/healthy", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok"))
	}))

	handler.HandleFunc("/i/pprof/", pprof.Index)
	handler.HandleFunc("/i/pprof/profile", pprof.Profile)
	for _, p := range rpprof.Profiles() {
		handler.Handle(fmt.Sprintf("/i/pprof/%s", p.Name()), pprof.Handler(p.Name()))
	}

	var grpcServer *grpc.Server
	if *grpcFlag != "" {
		log.Printf("Listening for GRPC on %s", *grpcFlag)
		grpcServer = grpc.NewServer(grpc.MaxRecvMsgSize(*grpcSizeFlag), grpc.MaxSendMsgSize(*grpcSizeFlag))
		pb.RegisterB6Server(grpcServer, b6grpc.NewB6Service(worlds, apiOptions, &lock))
		go func() {
			listener, err := net.Listen("tcp", *grpcFlag)
			if err == nil {
				err = grpcServer.Serve(listener)
			}
			if err != nil {
				os.Stdout.Write([]byte(err.Error()))
				os.Exit(1)
			}
		}()
	}

	server := http.Server{Addr: *httpFlag, Handler: handler}
	log.Printf("Listening for HTTP on %s", *httpFlag)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
