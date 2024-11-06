package utils

import (
	"encoding/csv"
	"encoding/xml"
	"net/http"
	"os"
)

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

func UpdateCSV(csvFile *os.File, w http.ResponseWriter, newRecords [][]string, csvWriter *csv.Writer) {
	err := csvFile.Truncate(0)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to truncate the metadata storage: ", err)
		return
	}

	_, err = csvFile.Seek(0, 0)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to move the cursor to the origin: ", err)
		return
	}

	for _, record := range newRecords {
		if err := csvWriter.Write(record); err != nil {
			DisplayError(w, http.StatusInternalServerError, "Failed to write to the metadata storage: ", err)
			return
		}
	}
}

func CheckBucketExistence(w http.ResponseWriter, bucketName, dir string) bool {
	bucketsCsv, err := os.Open(dir + "/buckets.csv")
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open the buckets.csv: ", err)
		return false
	}
	defer bucketsCsv.Close()

	bucketsCsvReader := csv.NewReader(bucketsCsv)
	bucketStorage, err := bucketsCsvReader.ReadAll()
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to read data from buckets.csv: ", err)
		return false
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
		return false
	} else {
		return true
	}
}

func CheckObjectExistence(w http.ResponseWriter, bucketName string, objectName, dir string) (bool, int, [][]string) {
	var objectID int
	var objectsRecords [][]string
	objectsCsv, err := os.OpenFile(dir+"/"+bucketName+"/objects.csv", os.O_RDWR, 0o644)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open objects.csv: ", err)
		return false, objectID, objectsRecords
	}
	defer objectsCsv.Close()

	objectsCsvReader := csv.NewReader(objectsCsv)
	objectsRecords, err = objectsCsvReader.ReadAll()
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to read data from objects.csv: ", err)
		return false, objectID, objectsRecords
	}

	// checking if an object exists in the objects.csv metadata storage
	var objectExistence bool
	for i, record := range objectsRecords {
		if record[0] == objectName {
			objectExistence = true
			objectID = i
			break
		}
	}

	// if an object does not exist - display an error
	if !objectExistence {
		DisplayErrorWoErr(w, http.StatusNotFound, "Such object does not exist")
		return false, objectID, objectsRecords
	}

	return true, objectID, objectsRecords
}

func CheckBucketExists(w http.ResponseWriter, bucketName, dir string) bool {
	bucketsCsv, err := os.Open(dir + "/buckets.csv")
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to open the buckets.csv: ", err)
		return false
	}
	defer bucketsCsv.Close()

	bucketsCsvReader := csv.NewReader(bucketsCsv)
	bucketStorage, err := bucketsCsvReader.ReadAll()
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to read data from buckets.csv: ", err)
		return false
	}

	// checking if a bucket exists in the buckets.csv metadata storage
	var bucketExistence bool
	for _, record := range bucketStorage {
		if record[0] == bucketName {
			bucketExistence = true
			return bucketExistence
		}
	}

	return bucketExistence
}
