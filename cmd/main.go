package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/iamsayantan/gimage"
	"github.com/iamsayantan/gimage/server"
)

const (
	defaultServingPort = "6050"
)

func main() {
	s3Config := gimage.S3Config{
		AWSKeyID:  "",
		AWSSecret: "",
		AWSRegion: "",
		S3Bucket:  "",
	}

	resizer := gimage.NewResizer()
	s3Uploader, err := gimage.NewS3Uploader(s3Config)

	if err != nil {
		panic(err.Error())
	}

	server := server.NewServer(resizer, s3Uploader)

	log.Printf("Starting HTTP server at port: %s", defaultServingPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", defaultServingPort), server))
}
