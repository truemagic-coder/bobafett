package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bevanhunt/gowrex"
	. "github.com/franela/goblin"
	"github.com/gin-gonic/gin"
	"github.com/prashantv/gostub"
	"github.com/rlmcpherson/s3gof3r"
)

func Test(t *testing.T) {
	g := Goblin(t)
	timeout := 10 * time.Second
	g.Describe("Request Specs:", func() {
		g.Before(func() {
			// set to release mode to hide debug warning
			gin.SetMode(gin.ReleaseMode)
			// disable logging in test
			log.SetOutput(ioutil.Discard)
		})
		g.Describe("/:", func() {
			g.It("route / should have proper body", func() {

				req, err := http.NewRequest("GET", "/", nil)
				if err != nil {
					fmt.Println(err)
				}
				resp := httptest.NewRecorder()
				// router and upload
				testRouter := GinEngine()
				testRouter.ServeHTTP(resp, req)
				g.Assert(resp.Code).Equal(200)
				g.Assert(resp.Body.String()).Equal("{\"url\":\"hello\"}\n")
			})
		})
		g.Describe("/upload:", func() {
			g.It("route /upload should return the 200/file on s3 upload success", func() {
				// stubs
				uuidStub := gostub.Stub(&uuidGenerator, func() string {
					return "test"
				})
				defer uuidStub.Reset()
				s3Stub := gostub.Stub(&s3Uploader, func(bucket *s3gof3r.Bucket, filename string, file multipart.File, c *gin.Context) error {
					return nil
				})
				defer s3Stub.Reset()

				extraParams := map[string]string{}
				req, err := gowrex.Request{
					URI:     "/upload",
					Timeout: timeout,
				}.PostFormFileDisk(extraParams, "file", "test.png")
				if err != nil {
					fmt.Println(err)
				}

				resp := httptest.NewRecorder()
				// router and upload
				testRouter := GinEngine()
				testRouter.ServeHTTP(resp, req.Req)
				// assert filename
				g.Assert(resp.Body.String()).Equal("{\"file\":\"test.png\"}\n")
				// assert 200
				g.Assert(resp.Code).Equal(200)
			})
			g.It("route /upload should return 500/error on s3 upload error", func() {
				// stubs
				uuidStub := gostub.Stub(&uuidGenerator, func() string {
					return "test"
				})
				defer uuidStub.Reset()
				s3Stub := gostub.Stub(&s3Uploader, func(bucket *s3gof3r.Bucket, filename string, file multipart.File, c *gin.Context) error {
					return errors.New("cannot upload to s3")
				})
				defer s3Stub.Reset()

				extraParams := map[string]string{}
				req, err := gowrex.Request{
					URI:     "/upload",
					Timeout: timeout,
				}.PostFormFileDisk(extraParams, "file", "test.png")
				if err != nil {
					fmt.Println(err)
				}

				resp := httptest.NewRecorder()
				// router and upload
				testRouter := GinEngine()
				testRouter.ServeHTTP(resp, req.Req)
				// assert error
				g.Assert(resp.Body.String()).Equal("{\"error\":\"there was an error uploading\"}\n")
				// assert 500
				g.Assert(resp.Code).Equal(500)
			})
			g.It("route /upload should return 400/error on no file", func() {
				extraParams := map[string]string{}
				req, err := gowrex.Request{
					URI:     "/upload",
					Timeout: timeout,
				}.PostFormFileDisk(extraParams, "missing_file", "test.png")
				if err != nil {
					fmt.Println(err)
				}

				resp := httptest.NewRecorder()
				// router and upload
				testRouter := GinEngine()
				testRouter.ServeHTTP(resp, req.Req)
				// assert error
				g.Assert(resp.Body.String()).Equal("{\"error\":\"you need to provide a file to upload\"}\n")
				// assert 500
				g.Assert(resp.Code).Equal(400)
			})
		})
		g.Describe("/download:", func() {
			g.It("route /download should return 400/error on no file", func() {
				req, err := gowrex.Request{
					URI:     "/download?filex=test.png",
					Timeout: timeout,
				}.Get()
				if err != nil {
					fmt.Println(err)
				}
				resp := httptest.NewRecorder()
				// router and upload
				testRouter := GinEngine()
				testRouter.ServeHTTP(resp, req.Req)
				// assert error
				g.Assert(resp.Body.String()).Equal("{\"error\":\"you must provide a file to download\"}\n")
				// assert 400
				g.Assert(resp.Code).Equal(400)
			})
			g.It("route /download should return 500/error on s3 error", func() {
				// stubs
				s3DownloadStub := gostub.Stub(&s3Downloader, func(bucket *s3gof3r.Bucket, filename string) (*bytes.Buffer, error) {
					return new(bytes.Buffer), errors.New("cannot download from s3")
				})
				defer s3DownloadStub.Reset()
				req, err := gowrex.Request{
					URI:     "/download?file=test.png",
					Timeout: timeout,
				}.Get()
				if err != nil {
					fmt.Println(err)
				}
				resp := httptest.NewRecorder()
				testRouter := GinEngine()
				testRouter.ServeHTTP(resp, req.Req)
				// assert error
				g.Assert(resp.Body.String()).Equal("{\"error\":\"there was an error downloading\"}\n")
				// assert 500
				g.Assert(resp.Code).Equal(500)
			})
			g.It("route /download should return 500/error on mimeType error", func() {
				// stubs
				s3DownloadStub := gostub.Stub(&s3Downloader, func(bucket *s3gof3r.Bucket, filename string) (*bytes.Buffer, error) {
					return new(bytes.Buffer), nil
				})
				defer s3DownloadStub.Reset()
				getMimeTypeStub := gostub.Stub(&getMimeTyper, func(s3Buffer *bytes.Buffer) (string, error) {
					return "", errors.New("mime type reading failed")
				})
				defer getMimeTypeStub.Reset()
				req, err := gowrex.Request{
					URI:     "/download?file=test.png",
					Timeout: timeout,
				}.Get()
				if err != nil {
					fmt.Println(err)
				}
				resp := httptest.NewRecorder()
				testRouter := GinEngine()
				testRouter.ServeHTTP(resp, req.Req)
				// assert error
				g.Assert(resp.Body.String()).Equal("{\"error\":\"there was an error reading the mime type\"}\n")
				// assert 500
				g.Assert(resp.Code).Equal(500)
			})
			g.It("route /download should return 200/file on success", func() {
				// stubs
				s3DownloadStub := gostub.Stub(&s3Downloader, func(bucket *s3gof3r.Bucket, filename string) (*bytes.Buffer, error) {
					return new(bytes.Buffer), nil
				})
				defer s3DownloadStub.Reset()
				getMimeTypeStub := gostub.Stub(&getMimeTyper, func(s3Buffer *bytes.Buffer) (string, error) {
					return "image/png", nil
				})
				defer getMimeTypeStub.Reset()
				req, err := gowrex.Request{
					URI:     "/download?file=test.png",
					Timeout: timeout,
				}.Get()
				if err != nil {
					fmt.Println(err)
				}
				resp := httptest.NewRecorder()
				testRouter := GinEngine()
				testRouter.ServeHTTP(resp, req.Req)
				// assert file stream
				g.Assert(resp.Body).Equal(new(bytes.Buffer))
				// assert 200
				g.Assert(resp.Code).Equal(200)
			})
		})
	})
}
