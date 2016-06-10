# Go S3 Microservice

## Local

### Setup
- ` brew update `
- ` brew install go `
- ` brew install glide `
- ` brew install libmagic `
-  setup [GOPATH](https://gist.github.com/vsouza/77e6b20520d07652ed7d) for your env file (.bashrc or .zshrc)
- ` git clone ` this repo into ` ~/golang/src `
- ` glide install ` in the local folder
- ` go get github.com/codegangsta/gin ` to install the reloading dev server

### Config
- rename `config.bak` to `config.json`
- add the proper keys - AWS_TOKEN is optional

### Run

#### Upload
- ` gin ` in the local folder then POST to ` localhost:3000/upload `  with a file to the `file` key
- will return either a 500 error or a 200 with s3 url (key) of saved file - the file is a uuid.file_extension

#### Download
- ` gin ` in the local folder then GET to ` localhost:3000/download/:key ` with a AWS filename (key) to the key param
- will return either a 500 error or a 200 with the file with the proper mime-type - streaming data

### Docker
- a docker file is provided that will build the project - you can use a config or set ENV VARS at your discretion

## Recommended Go Editor
- [Visual Studio Code](https://code.visualstudio.com/) with [Go Extension](https://github.com/Microsoft/vscode-go)
