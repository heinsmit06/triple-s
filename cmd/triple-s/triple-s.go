package s3

import (
	"fmt"
	"net/http"

	"triple-s/internal"
)

func Run() {
	router := http.NewServeMux()

	router.HandleFunc("POST /{BucketName}", internal.CreateBuckets)
	router.HandleFunc("GET /", internal.GetBuckets)
	router.HandleFunc("DELETE /{BucketName}", internal.DeleteBuckets)

	router.HandleFunc("POST /{BucketName}/{ObjectKey}", internal.CreateObjects)
	router.HandleFunc("GET /{BucketName}/{ObjectKey}", internal.GetObjects)
	router.HandleFunc("DELETE /{BucketName}/{ObjectKey}", internal.DeleteObjects)

	fmt.Println("Server is listening to :8080")
	http.ListenAndServe(":8080", router)
}
