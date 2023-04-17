package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	b6grpc "diagonal.works/b6/grpc"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
	pb "diagonal.works/b6/proto"

	"google.golang.org/grpc"
)

func main() {
	httpFlag := flag.String("http", "localhost:8100", "Host and port on which to serve HTTP")
	grpcFlag := flag.String("grpc", "localhost:8001", "Host and port on which to serve GRPC")
	grpcSizeFlag := flag.Int("grpc-size", 16*1024*1024, "Maximum size for GRPC messages")
	worldFlag := flag.String("world", "", "World to load")
	staticFlag := flag.String("static", "src/diagonal.works/b6/cmd/b6/js/static", "Path to static content")
	jsFlag := flag.String("js", "src/diagonal.works/b6/cmd/b6/js", "Path to JS bundle")
	coresFlag := flag.Int("cores", runtime.NumCPU(), "Number of cores available")
	flag.Parse()

	if *worldFlag == "" {
		os.Stdout.Write([]byte("Must specify --world\n"))
		os.Exit(1)
	}

	base, err := compact.ReadWorld(*worldFlag, *coresFlag)
	if err != nil {
		os.Stdout.Write([]byte(err.Error()))
		os.Exit(1)
	}
	w := ingest.NewMutableOverlayWorld(base)

	handler := http.NewServeMux()
	handler.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s", r.URL)
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(*staticFlag, "index.html"))
		} else {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}))
	handler.Handle("/b6.css", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(*staticFlag, "b6.css"))
	}))
	handler.Handle("/bundle.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(*jsFlag, "bundle.js"))
	}))

	handler.HandleFunc("/healthy", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok"))
	}))

	var grpcServer *grpc.Server
	var lock sync.RWMutex
	if *grpcFlag != "" {
		log.Printf("Listening for GRPC on %s", *grpcFlag)
		grpcServer = grpc.NewServer(grpc.MaxRecvMsgSize(*grpcSizeFlag), grpc.MaxSendMsgSize(*grpcSizeFlag))
		pb.RegisterB6Server(grpcServer, b6grpc.NewB6Service(w, *coresFlag, &lock))
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
		os.Stdout.Write([]byte(err.Error()))
		os.Exit(1)
	}
}
