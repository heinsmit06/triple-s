package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func CreateObjects(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path[1:]
	pathSlice := strings.Split(path, "/")
	bucketName := pathSlice[0]
	fmt.Println("path: " + path)
	fmt.Println("bucket name: " + bucketName)

	// checking whether a bucket exists or not
	_, err := os.Stat("data/" + bucketName)
	if os.IsNotExist(err) {
		DisplayError(w, http.StatusInternalServerError, "There is no such directory: ", err)
		return
	} else if err != nil {
		DisplayError(w, http.StatusNotFound, "Failed to check the directory presence: ", err)
		return
	}

	// reading request body
	objectContent, err := io.ReadAll(req.Body)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to read binary content of the request", err)
		return
	}

	// creating an object
	err = os.WriteFile("data/"+path, objectContent, 0o644)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to create a file", err)
		return
	}

	// creating objects.csv: it either appends or creates new entries
	objectsCsv, err := os.OpenFile("data/"+bucketName+"/objects.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Error creating objects.csv", err)
		return
	}
	defer objectsCsv.Close()

	// preparing object metada
	timeNow := time.Now().Format(time.RFC850)

	// detecting the content type based on request body
	contentType := http.DetectContentType(objectContent)
	fileInfo, err := os.Stat("data/" + path)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open the file: ", err)
		return
	}

	size := strconv.Itoa(int(fileInfo.Size()))
	record := []string{pathSlice[1], size, contentType, timeNow}

	// writing to objects.csv
	csv_writer := csv.NewWriter(objectsCsv)
	if err := csv_writer.Write(record); err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to write/append to objects.csv: ", err)
		return
	}
	defer csv_writer.Flush()

	DisplaySuccess(w, 200, "Object was created and metadata was written")
}

func GetObjects(w http.ResponseWriter, req *http.Request) {
}

func DeleteObjects(w http.ResponseWriter, req *http.Request) {
}
