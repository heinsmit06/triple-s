package internal

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"time"
)

func CreateBuckets(w http.ResponseWriter, req *http.Request) {
	// DONE. Bucket names must be unique across the system.
	// DONE. Names should be between 3 and 63 characters long.
	// DONE. Only lowercase letters, numbers, hyphens (-), and dots (.) are allowed.
	// Must not be formatted as an IP address (e.g., 192.168.0.1).
	// DONE. Must not begin or end with a hyphen and must not contain two consecutive periods or dashes.

	path := req.URL.Path[1:]
	fmt.Printf("Received request for path: '%s'\n", path)
	if path == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	} else if len(path) < 3 || len(path) > 63 {
		http.Error(w, "Incorrect bucket name length: must be between 3-63 chars", http.StatusBadRequest)
		return
	} else if path[0] == '-' || path[len(path)-1] == '-' {
		http.Error(w, "Must not begin or end with a hyphen", http.StatusBadRequest)
		return
	}

	for idx, ch := range path {
		if !((ch >= 97 && ch <= 122) || (ch >= 48 && ch <= 57) || (ch == '-') || (ch == '.')) {
			http.Error(w, "Forbidden rune is used in Bucket name", 400)
			return
		} else if idx <= (len(path) - 2) {
			if ch == '.' && path[idx+1] == '.' {
				http.Error(w, "Two consecutive periods are not allowed", http.StatusBadRequest)
				return
			} else if ch == '-' && path[idx+1] == '-' {
				http.Error(w, "Two consecutive dashes are not allowed", http.StatusBadRequest)
				return
			}
		}
	}

	_, errStat := os.Stat("data/" + path)
	if errStat == nil {
		http.Error(w, "Bucket already exists", http.StatusConflict)
		return
	} else if !os.IsNotExist(errStat) {
		http.Error(w, "Failed to check bucket existence: "+errStat.Error(), http.StatusInternalServerError)
		return
	}

	// Done with checking for errors, now creating the bucket and storing its metadata in a csv file
	err := os.MkdirAll("data/"+path, 0o755)
	if err != nil {
		panic(err)
	}

	// storing bucket metadata in metadata storage
	buckets_csv, ok := os.OpenFile("data/buckets.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if ok != nil {
		http.Error(w, "Error creating buckets.csv", http.StatusInternalServerError)
		return
	}
	defer buckets_csv.Close()

	// preparing the bucket metadata
	fileInfo, _ := os.Stat("data/" + path)
	time_now := time.Now().Local().Format(time.RFC850)
	bucket_field := []string{path, time_now, fileInfo.ModTime().Format(time.RFC850)}

	// writing the metadata into the metadata storage
	csv_writer := csv.NewWriter(buckets_csv)
	csv_err := csv_writer.Write(bucket_field)
	if csv_err != nil {
		http.Error(w, "Error writing to CSV metadata storage", http.StatusInternalServerError)
		return
	}

	// flushing the writer because writes are buffered and flush must be called in the end to actually write the record
	csv_writer.Flush()
	if csv_writer.Error() != nil {
		http.Error(w, "Error flushing the writer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Bucket was created and metadata is written\n")
}

func GetBuckets(w http.ResponseWriter, req *http.Request) {
	dirSlice, err := os.ReadDir("data/")
	if err != nil {
		http.Error(w, "Failed to list directories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, bucket := range dirSlice {
		fmt.Println(bucket.Name())
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "Buckets were listed\n")
}

func DeleteBuckets(w http.ResponseWriter, req *http.Request) {
}
