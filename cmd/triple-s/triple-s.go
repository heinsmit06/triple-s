package s3

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"triple-s/internal"
)

var helpMessage = `
Simple Storage Service.

**Usage:**
    triple-s [-port <N>] [-dir <S>]
    triple-s --help

	**Options:**
- --help     Show this screen.
- --port N   Port number
- --dir S    Path to the directory
`

func Run() {
	dirPtr := flag.String("dir", "data", "path to the directory where the files will be stored")
	portPtr := flag.String("port", "6666", "port value that the server will use")
	helpPtr := flag.Bool("help", false, "shows the usage information")

	flag.Parse()
	if *helpPtr {
		fmt.Println(helpMessage)
		return
	}
	portNumber, _ := strconv.Atoi(*portPtr)
	if portNumber < 1024 || portNumber > 65535 {
		fmt.Fprintf(os.Stderr, "Port number is not allowed: must be in between [1024, 65535]\n")
		return
	}

	router := http.NewServeMux()

	err := os.MkdirAll(*dirPtr, 0o755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create directory:", err)
		return
	}

	csvPath := *dirPtr + "/buckets.csv"
	file, err := os.Create(csvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create buckets.csv:", err)
		return
	}
	defer file.Close()

	router.HandleFunc("PUT /{BucketName}", func(w http.ResponseWriter, r *http.Request) {
		internal.CreateBuckets(w, r, *dirPtr)
	})
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		internal.GetBuckets(w, r, *dirPtr)
	})
	router.HandleFunc("DELETE /{BucketName}", func(w http.ResponseWriter, r *http.Request) {
		internal.DeleteBuckets(w, r, *dirPtr)
	})

	router.HandleFunc("PUT /{BucketName}/{ObjectKey}", func(w http.ResponseWriter, r *http.Request) {
		internal.CreateObjects(w, r, *dirPtr)
	})
	router.HandleFunc("GET /{BucketName}/{ObjectKey}", func(w http.ResponseWriter, r *http.Request) {
		internal.GetObjects(w, r, *dirPtr)
	})
	router.HandleFunc("DELETE /{BucketName}/{ObjectKey}", func(w http.ResponseWriter, r *http.Request) {
		internal.DeleteObjects(w, r, *dirPtr)
	})

	fmt.Println("Server is listening to: " + *portPtr)
	err = http.ListenAndServe(":"+*portPtr, router)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
	}
}
