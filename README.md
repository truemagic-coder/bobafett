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
- `gin` in the project folder

#### Upload
- POST to `localhost:3000/upload` with the `file` key and `folder` key
- will return a s3 url or an error

#### Download
- POST to `localhost:3000/download` with the `file` key and `folder` key
- will return a streamed file or an error

### Docker
- a docker file is provided that will build the project - you can use a config or set ENV VARS at your discretion

## Recommended Go Editor
- [Visual Studio Code](https://code.visualstudio.com/) with [Go Extension](https://github.com/Microsoft/vscode-go)

## TODO
- replace aws-sdk-go with s3gof3r
