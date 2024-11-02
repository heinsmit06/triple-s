package internal

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Object struct {
	XMLName          xml.Name `xml:"object"`
	ObjectKey        string
	Size             string
	ContentType      string
	LastModifiedTime string
}

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
	lastModifiedTime := time.Now().Format(time.RFC850)
	contentType := http.DetectContentType(objectContent)
	fileInfo, err := os.Stat("data/" + path)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open the file: ", err)
		return
	}
	size := strconv.Itoa(int(fileInfo.Size()))
	record := []string{pathSlice[1], size, contentType, lastModifiedTime}

	// checking if the same object was already in the .csv storage
	csvReader := csv.NewReader(objectsCsv)
	records, err := csvReader.ReadAll()
	alreadyPresent := false
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to read the metada from objects.csv: ", err)
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
		err = objectsCsv.Truncate(0)
		if err != nil {
			DisplayError(w, http.StatusInternalServerError, "Failed to truncate the metadata storage: ", err)
			return
		}

		_, err = objectsCsv.Seek(0, 0)
		if err != nil {
			DisplayError(w, http.StatusInternalServerError, "Failed to move the cursor to the origin: ", err)
			return
		}

		for _, record := range newRecords {
			if err := csv_writer.Write(record); err != nil {
				DisplayError(w, http.StatusInternalServerError, "Failed to write to the metadata storage: ", err)
				return
			}
		}
	} else {
		if err := csv_writer.Write(record); err != nil {
			DisplayError(w, http.StatusInternalServerError, "Failed to write/append to objects.csv: ", err)
			return
		}
	}

	DisplaySuccess(w, 200, "Object was created and metadata was written")
}

func GetObjects(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path[1:]
	pathSlice := strings.Split(path, "/")
	bucketName := pathSlice[0]
	fmt.Println("path: " + path)
	fmt.Println("bucket name: " + bucketName)

	bucketsCsv, err := os.Open("data/buckets.csv")
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open the buckets.csv: ", err)
		return
	}
	defer bucketsCsv.Close()

	bucketsCsvReader := csv.NewReader(bucketsCsv)
	bucketStorage, err := bucketsCsvReader.ReadAll()
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to read data from buckets.csv: ", err)
		return
	}

	// checking if a bucket exists in the buckets.csv metadata storage
	var bucketExistence bool
	for _, record := range bucketStorage {
		if record[0] == bucketName {
			bucketExistence = true
			break
		}
	}

	// if bucket does not exist - display an error
	if !bucketExistence {
		DisplayErrorWoErr(w, http.StatusNotFound, "No such bucket exists")
		return
	}

	// checking if an object exists in the objects.csv metadata storage
	objectsCsv, err := os.Open("data/" + bucketName + "/objects.csv")
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open objects.csv: ", err)
		return
	}
	defer objectsCsv.Close()

	objectsCsvReader := csv.NewReader(objectsCsv)
	objectsRecords, err := objectsCsvReader.ReadAll()
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to read data from objects.csv: ", err)
		return
	}

	// checking if an object exists in the objects.csv metadata storage
	var objectExistence bool
	var objectID int
	for i, record := range objectsRecords {
		if record[0] == pathSlice[1] {
			objectExistence = true
			objectID = i
			break
		}
	}

	// if an object does not exist - display an error
	if !objectExistence {
		DisplayErrorWoErr(w, http.StatusNotFound, "Such object does not exist")
		return
	}

	// opening the file to get its binary data
	file, err := os.Open("data/" + path)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open the file to read its binary data: ", err)
		return
	}
	defer file.Close()

	binaryFile, err := io.ReadAll(file)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to read binary content of the file: ", err)
		return
	}

	// setting headers and writing to http.ResponseWriter
	w.Header().Set("Content-Length", objectsRecords[objectID][1])
	w.Header().Set("Content-Type", objectsRecords[objectID][2])
	w.Header().Set("Last-Modified", objectsRecords[objectID][3])
	w.Write(binaryFile)
}

func DeleteObjects(w http.ResponseWriter, req *http.Request) {
}

// object := Object{
// 	ObjectKey:        objectsRecords[objectID][0],
// 	Size:             objectsRecords[objectID][1],
// 	ContentType:      objectsRecords[objectID][2],
// 	LastModifiedTime: objectsRecords[objectID][3],
// }

// out, err := xml.MarshalIndent(object, " ", "  ")
// if err != nil {
// 	DisplayError(w, 500, "Failed to encode XML", err)
// 	return
// }
// w.Write(out)
