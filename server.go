package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/rakyll/magicmime"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

func main() {
	// create gin server
	r := gin.Default()

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

	// upload route
	r.POST("/upload", func(c *gin.Context) {
		// get file upload
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Println("Failed to upload", err)
			c.JSON(400, gin.H{"error": "you need to provide a file to upload"})
		} else {
			// get key folder
			folder := c.PostForm("folder")
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
			filename := folder + u1.String() + fileExt
			// upload file to s3
			result, err := uploader.Upload(&s3manager.UploadInput{
				Body:   file,
				Bucket: aws.String(awsBucket),
				Key:    aws.String(filename),
			})
			// if can't download then throw error else return s3 key (url)
			if err != nil {
				log.Println("Failed to upload", err)
				c.JSON(500, gin.H{"error": "there was an error uploading"})
			} else {
				log.Println("Successfully uploaded to", result.Location)
				c.JSON(200, gin.H{"url": result.Location})
			}
		}
	})

	// download route
	r.POST("/download", func(c *gin.Context) {
		// get AWS key as param
		key := c.PostForm("file")
		folder := c.PostForm("folder")
		if key == "" {
			c.JSON(400, gin.H{"error": "you must provide a file to download"})
		} else {
			// setup file
			file, err := os.Create(key)
			if err != nil {
				c.JSON(500, gin.H{"error": "there was an error downloading"})
			} else {
				// close the file and delete after route call is done
				defer file.Close()
				defer os.Remove(key)
				// download file from s3
				downloader := s3manager.NewDownloader(session.New(&aws.Config{
					Credentials: credentials.NewStaticCredentials(awsID, awsSecret, awsToken),
					Region:      aws.String(awsRegion),
				}))
				filename := folder + key
				_, err = downloader.Download(file, &s3.GetObjectInput{
					Bucket: aws.String(awsBucket),
					Key:    aws.String(filename),
				})
				// if can't download from S3
				if err != nil {
					log.Println("Failed to download", err)
					c.JSON(500, gin.H{"error": "there was an error downloading"})
				} else {
					// init magicmime else throw error
					if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
						log.Println("Failed to read mime type", err)
						c.JSON(500, gin.H{"error": "there was an error reading the mime type"})
					} else {
						// close magicmime after route call is done
						defer magicmime.Close()
						// read file
						bytes, err := ioutil.ReadFile(key)
						// if can't read file then throw error
						if err != nil {
							log.Println("Failed to read file", err)
							c.JSON(500, gin.H{"error": "there was an error opening the file"})
						} else {
							// read mimetype from file buffer
							mimetype, err := magicmime.TypeByBuffer(bytes)
							// if can't read mimetype then throw error
							if err != nil {
								log.Println("Failed to read mime type", err)
								c.JSON(500, gin.H{"error": "there was an error reading the mime type"})
							} else {
								// stream data to the requestor
								c.Data(200, mimetype, bytes)
							}
						}
					}
				}
			}
		}
	})

	// run gin server
	r.Run()
}
