package main

import (
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
	// create gin server
	r := gin.Default()
	// create post file upload route
	r.POST("/upload", func(c *gin.Context) {
		// look for config file
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		err := viper.ReadInConfig()
		// if no config file then read from ENV VARS
		if err != nil {
			viper.AutomaticEnv()
		}
		// setup keys from config file or ENV VARS
		awsID := viper.GetString("AWS_ID")
		awsSecret := viper.GetString("AWS_SECRET")
		awsBucket := viper.GetString("AWS_BUCKET")
		awsRegion := viper.GetString("AWS_REGION")
		awsToken := viper.GetString("AWS_TOKEN")
		// get file upload
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Println("Failed to upload", err)
			c.JSON(400, gin.H{"error": "you need to provide a file to upload"})
		}
		// setup s3 uploader
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
		// upload file
		result, err := uploader.Upload(&s3manager.UploadInput{
			Body:   file,
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
