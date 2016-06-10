package main

import (
	"compress/gzip"
	"io"
	"log"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

func main() {
	// setup gin
	r := gin.Default()
	// create post file upload route
	r.POST("/upload", func(c *gin.Context) {
		// look for config file
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		err := viper.ReadInConfig()
		if err != nil {
			// setup keys from ENV VARS
			// TODO: test this
			viper.AutomaticEnv()
		}
		// setup keys from config file
		awsID := viper.GetString("AWS_ID")
		awsSecret := viper.GetString("AWS_SECRET")
		awsBucket := viper.GetString("AWS_BUCKET")
		awsRegion := viper.GetString("AWS_REGION")
		awsToken := viper.GetString("AWS_TOKEN")
		// get file upload
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Println("Failed to upload", err)
			c.JSON(400, gin.H{"error": err})
		}
		// gzip stream file contents
		reader, writer := io.Pipe()
		go func() {
			gw := gzip.NewWriter(writer)
			io.Copy(gw, file)

			file.Close()
			gw.Close()
			writer.Close()
		}()
		uploader := s3manager.NewUploader(session.New(&aws.Config{
			Credentials: credentials.NewStaticCredentials(awsID, awsSecret, awsToken),
			Region:      aws.String(awsRegion),
		}))
		// create uuid v4
		u1 := uuid.NewV4()
		// get file extension
		fileExt := filepath.Ext(header.Filename)
		// create unique filename
		filename := u1.String() + fileExt
		// upload file from gzip streaming file
		result, err := uploader.Upload(&s3manager.UploadInput{
			Body:   reader,
			Bucket: aws.String(awsBucket),
			Key:    aws.String(filename),
		})
		// if error 400 with error else 200 with s3 url
		if err != nil {
			log.Println("Failed to upload", err)
			c.JSON(400, gin.H{"error": "there was an error uploading"})
		} else {
			log.Println("Successfully uploaded to", result.Location)
			c.JSON(200, gin.H{"url": result.Location})
		}
	})
	// run gin server
	r.Run()
}
