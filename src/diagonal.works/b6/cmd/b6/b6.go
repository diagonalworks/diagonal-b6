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
	coresFlag := flag.Int("cores", runtime.NumCPU(), "Number of cores available")
	fileIOFlag := flag.Bool("file-io", true, "Is file IO allowed from the API?")
	flag.Parse()

	if *worldFlag == "" {
		fmt.Fprintln(os.Stderr, "Must specify --world")
		os.Exit(1)
	}

	base, err := compact.ReadWorld(*worldFlag, *coresFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	var w ingest.MutableWorld
	if *readOnlyFlag {
		w = ingest.ReadOnlyWorld{World: base}
	} else {
		w = ingest.NewMutableOverlayWorld(base)
	}

	apiOptions := api.Options{
		Cores:         *coresFlag,
		FileIOAllowed: *fileIOFlag,
	}

	options := ui.Options{
		StaticPath:     *staticFlag,
		JavaScriptPath: *jsFlag,
		World:          w,
		APIOptions:     apiOptions,
	}

	handler := http.NewServeMux()
	if err := ui.RegisterWebInterface(handler, &options); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	ui.RegisterTiles(handler, w, *coresFlag)

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
	var lock sync.RWMutex
	if *grpcFlag != "" {
		log.Printf("Listening for GRPC on %s", *grpcFlag)
		grpcServer = grpc.NewServer(grpc.MaxRecvMsgSize(*grpcSizeFlag), grpc.MaxSendMsgSize(*grpcSizeFlag))
		pb.RegisterB6Server(grpcServer, b6grpc.NewB6Service(w, apiOptions, &lock))
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
