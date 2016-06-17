package main

import (
	"bytes"
	"io"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rakyll/magicmime"
	"github.com/rlmcpherson/s3gof3r"
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
	awsID := viper.GetString("AWS_ACCESS_KEY_ID")
	awsSecret := viper.GetString("AWS_SECRET_ACCESS_KEY")
	s3Bucket := viper.GetString("S3_BUCKET")
	s3Domain := viper.GetString("S3_DOMAIN")
	awsToken := viper.GetString("AWS_SECURITY_TOKEN")

	keys := s3gof3r.Keys{AccessKey: awsID, SecretKey: awsSecret, SecurityToken: awsToken}
	s3 := s3gof3r.S3{Keys: keys, Domain: s3Domain}
	bucket := s3.Bucket(s3Bucket)

	// upload route
	r.POST("/upload", func(c *gin.Context) {
		// get file upload
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Println("Failed to upload", err)
			c.JSON(400, gin.H{"error": "you need to provide a file to upload"})
			return
		}
		// get key folder
		folder := c.PostForm("folder")
		// create uuid v4
		u1 := uuid.NewV4()
		// get file extension
		fileExt := filepath.Ext(header.Filename)
		// create unique filename and folder as well
		filename := folder + u1.String() + fileExt
		// Open a PutWriter for upload
		w, err := bucket.PutWriter(filename, nil, nil)
		if err != nil {
			log.Println("Failed to upload", err)
			c.JSON(500, gin.H{"error": "there was an error uploading"})
			return
		}
		if _, err = io.Copy(w, file); err != nil { // Copy into S3
			log.Println("Failed to upload", err)
			c.JSON(500, gin.H{"error": "there was an error uploading"})
			return
		}
		if err = w.Close(); err != nil {
			log.Println("Failed to upload", err)
			c.JSON(500, gin.H{"error": "there was an error uploading"})
			return
		}
		log.Println("Successfully uploaded to", filename)
		c.JSON(200, gin.H{"file": filename})
		return
	})

	// download route
	r.POST("/download", func(c *gin.Context) {
		// get AWS key as param
		key := c.PostForm("file")
		folder := c.PostForm("folder")
		if key == "" {
			c.JSON(400, gin.H{"error": "you must provide a file to download"})
			return
		}
		// combine folder and key
		filename := folder + key
		r, _, err := bucket.GetReader(filename, nil)
		defer r.Close()
		// if can't download from S3
		if err != nil {
			log.Println("Failed to download", err)
			c.JSON(500, gin.H{"error": "there was an error downloading"})
			return
		}
		// stream to bytes buffer
		s3Buffer := new(bytes.Buffer)
		if _, err = io.Copy(s3Buffer, r); err != nil {
			log.Println("Failed to download", err)
			c.JSON(500, gin.H{"error": "there was an error downloading"})
			return
		}
		// init magicmime else throw error
		if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
			log.Println("Failed to read mime type", err)
			c.JSON(500, gin.H{"error": "there was an error reading the mime type"})
			return
		}
		// close magicmime after route call is done
		defer magicmime.Close()
		// read mimetype from file buffer
		mimetype, err := magicmime.TypeByBuffer(s3Buffer.Bytes())
		// if can't read mimetype then throw error
		if err != nil {
			log.Println("Failed to read mime type", err)
			c.JSON(500, gin.H{"error": "there was an error reading the mime type"})
			return
		}
		// stream data to the requestor
		c.Data(200, mimetype, s3Buffer.Bytes())
		return
	})

	// run gin server
	r.Run()
}
