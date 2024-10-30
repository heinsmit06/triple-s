package internal

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Bucket struct {
	XMLName      xml.Name `xml:"Bucket"`
	Name         string   `xml:"Name"`
	CreationTime string   `xml:"CreationTime"`
	LastModTime  string   `xml:"LastModifiedTime"`
	IsEmpty      string
}

type Buckets struct {
	XMLName xml.Name `xml:"Buckets"`
	Buckets []Bucket
}

type ErrorResponse struct {
	XMLName    xml.Name `xml:"error"`
	StatusCode int
	Message    string
}

type SuccessResponse struct {
	XMLName    xml.Name `xml:"success"`
	StatusCode int
	Message    string
}

func DisplayError(w http.ResponseWriter, statusCode int, message string, err error) {
	errorResponse := ErrorResponse{
		StatusCode: statusCode,
		Message:    message + err.Error(),
	}
	out, _ := xml.MarshalIndent(errorResponse, " ", "  ")
	w.Header().Set("Content-type", "application/xml")
	w.WriteHeader(statusCode)
	w.Write(out)
}

func DisplayErrorWoErr(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := ErrorResponse{
		StatusCode: statusCode,
		Message:    message,
	}
	out, _ := xml.MarshalIndent(errorResponse, " ", "  ")
	w.Header().Set("Content-type", "application/xml")
	w.WriteHeader(statusCode)
	w.Write(out)
}

func DisplaySuccess(w http.ResponseWriter, statusCode int, message string) {
	successResponse := SuccessResponse{
		StatusCode: statusCode,
		Message:    message,
	}
	out, _ := xml.MarshalIndent(successResponse, " ", "  ")
	w.Header().Set("Content-type", "application/xml")
	w.WriteHeader(statusCode)
	w.Write(out)
}

func GetBuckets(w http.ResponseWriter, req *http.Request) {
	buckets_csv, err := os.Open("data/buckets.csv")
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open buckets.csv: ", err)
		return
	}
	defer buckets_csv.Close()

	reader := csv.NewReader(buckets_csv)
	records, err := reader.ReadAll()
	if err != nil {
		DisplayError(w, 500, "Failed to parse metadata of buckets", err)
		return
	}

	var buckets []Bucket
	for _, record := range records {
		if len(record) < 3 {
			continue
		}
		bucket := Bucket{
			Name:         record[0],
			CreationTime: record[1],
			LastModTime:  record[2],
		}
		buckets = append(buckets, bucket)
	}

	response := Buckets{
		Buckets: buckets,
	}
	w.Header().Set("Content-type", "application/xml")
	out, err := xml.MarshalIndent(response, " ", "  ")
	if err != nil {
		DisplayError(w, 500, "Failed to encode XML", err)
		return
	}
	w.Write(out)
}

func CreateBuckets(w http.ResponseWriter, req *http.Request) {
	// DONE. Bucket names must be unique across the system.
	// DONE. Names should be between 3 and 63 characters long.
	// DONE. Only lowercase letters, numbers, hyphens (-), and dots (.) are allowed.
	// Must not be formatted as an IP address (e.g., 192.168.0.1).
	// DONE. Must not begin or end with a hyphen and must not contain two consecutive periods or dashes.

	path := req.URL.Path[1:]
	fmt.Printf("Received request for path: '%s'\n", path)
	if path == "" {
		DisplayErrorWoErr(w, http.StatusBadRequest, "Bucket name is required")
		return
	} else if len(path) < 3 || len(path) > 63 {
		DisplayErrorWoErr(w, http.StatusBadRequest, "Incorrect bucket name length: must be between 3-63 chars")
		return
	} else if path[0] == '-' || path[len(path)-1] == '-' {
		DisplayErrorWoErr(w, http.StatusBadRequest, "Must not begin or end with a hyphen")
		return
	}

	for idx, ch := range path {
		if !((ch >= 97 && ch <= 122) || (ch >= 48 && ch <= 57) || (ch == '-') || (ch == '.')) {
			DisplayErrorWoErr(w, http.StatusBadRequest, "Forbidden rune is used in Bucket name")
			return
		} else if idx <= (len(path) - 2) {
			if ch == '.' && path[idx+1] == '.' {
				DisplayErrorWoErr(w, http.StatusBadRequest, "Two consecutive periods are not allowed")
				return
			} else if ch == '-' && path[idx+1] == '-' {
				DisplayErrorWoErr(w, http.StatusBadRequest, "Two consecutive dashes are not allowed")
				return
			}
		}
	}

	_, errStat := os.Stat("data/" + path)
	if errStat == nil {
		DisplayErrorWoErr(w, http.StatusConflict, "Bucket already exists")
		return
	} else if !os.IsNotExist(errStat) {
		DisplayError(w, http.StatusInternalServerError, "Failed to check bucket existence: ", errStat)
		return
	}

	// done with checking for errors, now creating the bucket and storing its metadata in a csv file
	err := os.MkdirAll("data/"+path, 0o755)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to create a bucket", err)
		return
	}

	// storing bucket metadata in metadata storage
	buckets_csv, ok := os.OpenFile("data/buckets.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if ok != nil {
		DisplayError(w, http.StatusInternalServerError, "Error creating buckets.csv", ok)
		return
	}
	defer buckets_csv.Close()

	// preparing the bucket metadata
	time_now := time.Now().Format(time.RFC850)
	bucket_field := []string{path, time_now, time_now, "True"} // bucket name, creation time, last modified time, emptiness of a bucket

	// writing the metadata into the metadata storage
	csv_writer := csv.NewWriter(buckets_csv)
	csv_err := csv_writer.Write(bucket_field)
	if csv_err != nil {
		DisplayError(w, http.StatusInternalServerError, "Error writing to CSV metadata storage", csv_err)
		return
	}

	// flushing the writer because writes are buffered and flush must be called in the end to actually write the record
	defer csv_writer.Flush()
	if csv_writer.Error() != nil {
		DisplayError(w, http.StatusInternalServerError, "Error flushing the writer", csv_writer.Error())
		return
	}

	DisplaySuccess(w, http.StatusOK, "Bucket was created and metadata is written")
}

func DeleteBuckets(w http.ResponseWriter, req *http.Request) {
	// getting a path from http.Request and opening csv storage to read the data from
	path := req.URL.Path[1:]
	buckets_csv, err := os.OpenFile("data/buckets.csv", os.O_RDWR, 0o644)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open buckets.csv", err)
		return
	}
	defer buckets_csv.Close()

	// creating a new csv reader and ReadAll of its rows
	csv_reader := csv.NewReader(buckets_csv)
	records, err := csv_reader.ReadAll()
	if err != nil {
		DisplayError(w, 500, "Failed to parse metadata of buckets", err)
		return
	}

	var updatedRecords [][]string
	var present, empty bool
	// checking each record if the bucket name is equal to bucket name for deletion
	for _, record := range records {
		if record[0] != path {
			updatedRecords = append(updatedRecords, record)
		} else if record[0] == path && record[3] == "True" {
			present = true
			empty = true
		} else if record[0] == path && record[3] == "False" {
			present = true
			empty = false
		}
	}

	if present && empty {
		err := os.Remove("data/" + path)
		if err != nil {
			DisplayError(w, http.StatusInternalServerError, "Failed to delete the bucket: ", err)
			return
		} else {
			DisplaySuccess(w, http.StatusOK, "Successfully deleted the bucket")
			// erasing all of its contents
			err = buckets_csv.Truncate(0)
			if err != nil {
				DisplayError(w, http.StatusInternalServerError, "Failed to truncate the metadata storage: ", err)
				return
			}

			_, err = buckets_csv.Seek(0, 0)
			if err != nil {
				DisplayError(w, http.StatusInternalServerError, "Failed to move the cursor to the origin: ", err)
				return
			}

			csv_writer := csv.NewWriter(buckets_csv)
			defer csv_writer.Flush()

			for _, record := range updatedRecords {
				if err := csv_writer.Write(record); err != nil {
					DisplayError(w, http.StatusInternalServerError, "Failed to write to the metadata storage: ", err)
					return
				}
			}
		}
	} else if present && !empty {
		DisplayErrorWoErr(w, http.StatusMethodNotAllowed, "Failed to delete the bucket - it is not empty")
		return
	} else if !present {
		DisplayErrorWoErr(w, http.StatusNotFound, "There is no such bucket")
		return
	}
}
