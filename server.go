package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rakyll/magicmime"
	"github.com/rlmcpherson/s3gof3r"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

// assign function to memory for testing
var s3Downloader = s3Download

// download from s3
func s3Download(bucket *s3gof3r.Bucket, filename string) (*bytes.Buffer, error) {
	// stream to bytes buffer
	s3Buffer := new(bytes.Buffer)
	// download file from s3
	r, _, err := bucket.GetReader(filename, nil)
	// if can't download from S3
	if err != nil {
		return nil, err
	}
	// copy s3 download into buffer
	if _, err = io.Copy(s3Buffer, r); err != nil {
		return nil, err
	}
	return s3Buffer, nil
}

func s3DownloadErr(err error, c *gin.Context) {
	log.Println("Failed to download", err)
	c.JSON(500, gin.H{"error": "there was an error downloading"})
	return
}

func mimeTypeErr(err error, c *gin.Context) {
	log.Println("Failed to read mime type", err)
	c.JSON(500, gin.H{"error": "there was an error reading the mime type"})
	return
}

var getMimeTyper = getMimeType

func getMimeType(s3Buffer *bytes.Buffer) (string, error) {
	// init magicmime else throw error
	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
		return "", err
	}
	// close magicmime after route call is done
	defer magicmime.Close()
	// read mimetype from file buffer
	return magicmime.TypeByBuffer(s3Buffer.Bytes())
}

func s3UploadErr(err error, c *gin.Context) {
	log.Println("Failed to upload", err)
	c.JSON(500, gin.H{"error": "there was an error uploading"})
	return
}

var s3Uploader = s3Upload

func s3Upload(bucket *s3gof3r.Bucket, filename string, file multipart.File, c *gin.Context) error {
	w, err := bucket.PutWriter(filename, nil, nil)
	if _, err = io.Copy(w, file); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return nil
}

var uuidGenerator = uuidGenerate

func uuidGenerate() string {
	u1 := uuid.NewV4()
	return u1.String()
}

func createUploadFilename(header *multipart.FileHeader, folder string) string {
	fileExt := filepath.Ext(header.Filename)
	return folder + uuidGenerator() + fileExt
}

// GinEngine is gin router.
func GinEngine() *gin.Engine {
	r := gin.New()

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

	// route to ping
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"url": "hello"})
		return
	})

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
		// create filename
		filename := createUploadFilename(header, folder)
		// upload file to S3
		if err := s3Uploader(bucket, filename, file, c); err != nil {
			s3UploadErr(err, c)
			return
		}
		// success
		log.Println("Successfully uploaded to", filename)
		c.JSON(200, gin.H{"file": filename})
		return
	})

	// download route
	r.GET("/download", func(c *gin.Context) {
		// get AWS key as param
		key, ok := c.GetQuery("file")
		if ok == false {
			c.JSON(400, gin.H{"error": "you must provide a file to download"})
			return
		}
		key, err := url.QueryUnescape(key)
		if err != nil {
			c.JSON(500, gin.H{"error": "file key could not be unescaped"})
			return
		}
		s3Buffer, err := s3Downloader(bucket, key)
		if err != nil {
			s3DownloadErr(err, c)
			return
		}
		mimeType, err := getMimeTyper(s3Buffer)
		// if can't read mimetype then throw error
		if err != nil {
			mimeTypeErr(err, c)
			return
		}
		// stream data to the requestor
		c.Data(200, mimeType, s3Buffer.Bytes())
		return
	})

	return r
}

func main() {
	GinEngine().Run()
}
