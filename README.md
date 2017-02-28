# Bobafett - Go S3 Microservice

[![Go Report Card](https://goreportcard.com/badge/github.com/bevanhunt/bobafett)](https://goreportcard.com/report/github.com/bevanhunt/bobafett)
[![GoDoc](https://godoc.org/github.com/bevanhunt/bobafett?status.svg)](https://godoc.org/github.com/bevanhunt/bobafett)
[![Build Status](https://img.shields.io/travis/bevanhunt/bobafett/master.svg)](https://travis-ci.org/bevanhunt/bobafett)
[![CodeCov](https://img.shields.io/codecov/c/github/bevanhunt/bobafett/master.svg)](https://codecov.io/gh/bevanhunt/bobafett/branch/master)

## Local

### Setup
- ` brew update `
- ` brew install go `
- ` brew install glide `
- ` brew install libmagic `
-  setup GOPATH for your env file (.bashrc or .zshrc):
```bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```
- ` git clone ` this repo into ` ~/go/src `
- ` glide install ` in the local folder
- ` go get github.com/codegangsta/gin ` to install the reloading dev server

### Config
- rename `config.bak` to `config.json`
- add the proper keys - AWS_SECRET_TOKEN is optional

### Run
- `gin` in the project folder

#### Upload
- POST to `localhost:3000/upload` with the `file` key and `folder` key
- will return a s3 url or an error

#### Download
- GET to `localhost:3000/download` with the `file` param encoded
- will return a streamed file or an error

#### Tests
- `go test` in the project folder

### Docker
- a docker file is provided that will build the project - you can use a config or set ENV VARS at your discretion

## Recommended Go Editor
- [Visual Studio Code](https://code.visualstudio.com/) with [Go Extension](https://github.com/Microsoft/vscode-go)
