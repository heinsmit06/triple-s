package s3

import (
	"fmt"
	"net/http"
	"os"
)

func Run() {
	router := http.NewServeMux()

	// router.HandleFunc("/{BucketName}", createBucket)
	router.HandleFunc("/", headers)
	// http.HandleFunc("/headers", headers)

	router.HandleFunc("POST /{bucketName}", createBucket)

	fmt.Println("Server is listening to :8080")
	http.ListenAndServe(":8080", router)
}

func createBucket(w http.ResponseWriter, req *http.Request) {
	// unnecessary check, because everything other than POST method will not be even considered in this case
	// if req.Method != http.MethodPost {
	// 	http.Error(w, "Method not allowed 123", http.StatusMethodNotAllowed)
	// 	return
	// }

	path := req.URL.Path[1:]
	fmt.Printf("Received request for path: '%s'\n", path)
	if path == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	_, errStat := os.Stat("data/" + path)
	if errStat == nil {
		http.Error(w, "Bucket already exists", http.StatusConflict)
		return
	} else if !os.IsNotExist(errStat) {
		http.Error(w, "Failed to check bucket existence: "+errStat.Error(), http.StatusInternalServerError)
		return
	}

	err := os.MkdirAll("data/"+path, 0o755)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Bucket was created\n")
}

func headers(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "RR is the ultimate car\n")
	// for name, headers := range req.Header {
	// 	for _, h := range headers {
	// 		fmt.Fprintf(w, "%v: %v\n", name, h)
	// 	}
	// }
}
