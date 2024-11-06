package internal

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"triple-s/utils"
)

type Object struct {
	XMLName          xml.Name `xml:"object"`
	ObjectKey        string
	Size             string
	ContentType      string
	LastModifiedTime string
}

func CreateObjects(w http.ResponseWriter, req *http.Request, dir string) {
	path := req.URL.Path[1:]
	pathSlice := strings.Split(path, "/")
	fmt.Println(dir)

	if len(pathSlice) < 2 {
		utils.DisplayErrorWoErr(w, http.StatusBadRequest, "Invalid path format")
		return
	}

	bucketName := pathSlice[0]
	objectKey := pathSlice[1]

	fmt.Println("path:", path)
	fmt.Println("bucket name:", bucketName)
	fmt.Println("object key:", objectKey)

	pattern := `^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]*[a-z0-9])?)*$`
	r, err := regexp.Compile(pattern)
	if err != nil {
		utils.DisplayErrorWoErr(w, http.StatusInternalServerError, "Failed to compile regex pattern")
		return
	}

	if !r.MatchString(objectKey) {
		utils.DisplayErrorWoErr(w, http.StatusBadRequest, "Incorrect object name")
		return
	}

	// checking whether a bucket exists or not
	bucketExistence := utils.CheckBucketExistence(w, bucketName, dir)
	if !bucketExistence {
		return
	}

	// reading request body
	objectContent, err := io.ReadAll(req.Body)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to read binary content of the request", err)
		return
	}

	// creating an object
	err = os.WriteFile(dir+"/"+path, objectContent, 0o644)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to create a file", err)
		return
	}

	// creating objects.csv: it either appends or creates new entries
	objectsCsv, err := os.OpenFile(dir+"/"+bucketName+"/objects.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Error creating objects.csv", err)
		return
	}
	defer objectsCsv.Close()

	// preparing object metada
	lastModifiedTime := time.Now().Format(time.RFC850)
	contentType := http.DetectContentType(objectContent)
	fileInfo, err := os.Stat(dir + "/" + path)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to open the file: ", err)
		return
	}
	size := strconv.Itoa(int(fileInfo.Size()))
	record := []string{pathSlice[1], size, contentType, lastModifiedTime}

	// checking if the same object was already in the .csv storage
	alreadyPresent := false
	csvReader := csv.NewReader(objectsCsv)
	records, err := csvReader.ReadAll()
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to read the metada from objects.csv: ", err)
		return
	}

	var newRecords [][]string
	for _, v := range records {
		if v[0] == pathSlice[1] {
			alreadyPresent = true
			newRecords = append(newRecords, record)
		} else {
			newRecords = append(newRecords, v)
		}
	}

	// depending on whether the object was already in the bucket:
	// either adding metadata of a new object
	// or updating metadata of an existing object

	// writing to objects.csv
	csv_writer := csv.NewWriter(objectsCsv)
	defer csv_writer.Flush()
	if alreadyPresent {
		utils.UpdateCSV(objectsCsv, w, newRecords, csv_writer)
	} else {
		if err := csv_writer.Write(record); err != nil {
			utils.DisplayError(w, http.StatusInternalServerError, "Failed to write/append to objects.csv: ", err)
			return
		}
	}

	// setting the status of a bucket as Active in buckets.csv
	// and updating its last modified time
	bucketsCsv, err := os.OpenFile(dir+"/buckets.csv", os.O_RDWR, 0o644)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to open buckets.csv: ", err)
		return
	}
	defer bucketsCsv.Close()

	csvReaderBuckets := csv.NewReader(bucketsCsv)
	bucketRecords, err := csvReaderBuckets.ReadAll()
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to read the buckets.csv: ", err)
		return
	}

	var newBucketRecords [][]string
	for _, bucketRecord := range bucketRecords {
		if bucketRecord[0] == bucketName {
			newBucketRecord := []string{bucketName, bucketRecord[1], time.Now().Format(time.RFC850), "False"}
			newBucketRecords = append(newBucketRecords, newBucketRecord)
		} else {
			newBucketRecords = append(newBucketRecords, bucketRecord)
		}
	}

	// writing to the buckets.csv the lastModifiedTime and Empty/Not Empty
	csvWriterBuckets := csv.NewWriter(bucketsCsv)
	defer csvWriterBuckets.Flush()
	utils.UpdateCSV(bucketsCsv, w, newBucketRecords, csvWriterBuckets)

	utils.DisplaySuccess(w, 200, "Object was created and metadata was written")
}

func GetObjects(w http.ResponseWriter, req *http.Request, dir string) {
	path := req.URL.Path[1:]
	pathSlice := strings.Split(path, "/")
	bucketName := pathSlice[0]
	fmt.Println("path: " + path)
	fmt.Println("bucket name: " + bucketName)

	// checking bucket existence
	bucketExistence := utils.CheckBucketExistence(w, bucketName, dir)
	if !bucketExistence {
		return
	}

	// checking if an object exists in the objects.csv metadata storage
	objectExistence, objectID, objectsRecords := utils.CheckObjectExistence(w, bucketName, pathSlice[1], dir)
	if !objectExistence {
		return
	}

	// opening the file to get its binary data
	file, err := os.Open(dir + "/" + path)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to open the file to read its binary data: ", err)
		return
	}
	defer file.Close()

	binaryFile, err := io.ReadAll(file)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to read binary content of the file: ", err)
		return
	}

	// setting headers and writing to http.ResponseWriter
	w.Header().Set("Content-Length", objectsRecords[objectID][1])
	w.Header().Set("Content-Type", objectsRecords[objectID][2])
	w.Header().Set("Last-Modified", objectsRecords[objectID][3])
	w.Write(binaryFile)
}

func DeleteObjects(w http.ResponseWriter, req *http.Request, dir string) {
	path := req.URL.Path[1:]
	pathSlice := strings.Split(path, "/")
	bucketName := pathSlice[0]
	fmt.Println("path: " + path)
	fmt.Println("bucket name: " + bucketName)

	bucketExistence := utils.CheckBucketExistence(w, bucketName, dir)
	if !bucketExistence {
		return
	}

	objectExistence, _, objectsRecords := utils.CheckObjectExistence(w, bucketName, pathSlice[1], dir)
	if !objectExistence {
		return
	}

	objectsCsv, err := os.OpenFile(dir+"/"+bucketName+"/objects.csv", os.O_RDWR, 0o644)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to open the objects.csv: ", err)
		return
	}
	defer objectsCsv.Close()

	// deleting the object
	err = os.Remove(dir + "/" + bucketName + "/" + pathSlice[1])
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to delete the file: ", err)
		return
	}

	// deleting its metadata
	var newObjectRecords [][]string
	for _, objectRecord := range objectsRecords {
		if objectRecord[0] == pathSlice[1] {
			continue
		} else {
			newObjectRecords = append(newObjectRecords, objectRecord)
		}
	}

	csvWriterObjects := csv.NewWriter(objectsCsv)
	utils.UpdateCSV(objectsCsv, w, newObjectRecords, csvWriterObjects)
	csvWriterObjects.Flush()

	// checking if objects.csv is empty and updating the bucket's status
	_, err = objectsCsv.Seek(0, 0)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to reset file pointer: ", err)
		return
	}

	csvReaderObjects := csv.NewReader(objectsCsv)
	csvObjectRecords, err := csvReaderObjects.ReadAll()
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to read from objects.csv: ", err)
		return
	}

	bucketIsEmpty := "False"
	if len(csvObjectRecords) == 0 {
		bucketIsEmpty = "True"
	}
	fmt.Println("len of csvObjectRecords: " + strconv.Itoa(len(csvObjectRecords)))

	// updating the last modified time of a bucket in buckets.csv
	bucketsCsv, err := os.OpenFile(dir+"/buckets.csv", os.O_RDWR, 0o644)
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to open buckets.csv: ", err)
		return
	}
	defer bucketsCsv.Close()

	csvReaderBuckets := csv.NewReader(bucketsCsv)
	bucketRecords, err := csvReaderBuckets.ReadAll()
	if err != nil {
		utils.DisplayError(w, http.StatusInternalServerError, "Failed to read the buckets.csv: ", err)
		return
	}

	var newBucketRecords [][]string
	for _, bucketRecord := range bucketRecords {
		if bucketRecord[0] == bucketName {
			newBucketRecord := []string{bucketName, bucketRecord[1], time.Now().Format(time.RFC850), bucketIsEmpty}
			newBucketRecords = append(newBucketRecords, newBucketRecord)
		} else {
			newBucketRecords = append(newBucketRecords, bucketRecord)
		}
	}

	// writing to the buckets.csv the lastModifiedTime and IsEmpty
	csvWriterBuckets := csv.NewWriter(bucketsCsv)
	defer csvWriterBuckets.Flush()
	utils.UpdateCSV(bucketsCsv, w, newBucketRecords, csvWriterBuckets)
	w.WriteHeader(http.StatusNoContent)
}
