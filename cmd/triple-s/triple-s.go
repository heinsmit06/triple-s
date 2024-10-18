package s3

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func Run() {
	fmt.Println("triple-s was started")

	router := http.NewServeMux()

	router.HandleFunc("/", createBucket)
	// http.HandleFunc("/headers", headers)

	http.ListenAndServe(":8080", nil)
}

func createBucket(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract bucket name from the URL path (e.g., /kasl)
	path := strings.TrimPrefix(req.URL.Path, "/")
	if path == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	err := os.MkdirAll("data/"+path, 0o755)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "Bucket was created\n")
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}
