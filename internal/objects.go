package internal

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func CreateObjects(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path[1:]
	pathSlice := strings.Split(path, "/")
	bucketName := pathSlice[0]
	fmt.Println("path: " + path)
	fmt.Println("bucket name: " + bucketName)

	_, err := os.Stat("data/" + bucketName)
	if err == nil {
	} else if os.IsNotExist(err) {
		DisplayError(w, http.StatusInternalServerError, "There is no such directory: ", err)
		return
	} else {
		DisplayError(w, http.StatusNotFound, "Failed to check the directory presence: ", err)
		return
	}

	object, err := os.OpenFile("data/"+path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		DisplayError(w, http.StatusInternalServerError, "Failed to create/open a file", err)
		return
	}
	defer object.Close()
}

func GetObjects(w http.ResponseWriter, req *http.Request) {
}

func DeleteObjects(w http.ResponseWriter, req *http.Request) {
}
