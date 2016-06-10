# Go S3 Microservice

## Local Setup
- ` brew update `
- ` brew install go `
- ` brew install glide ` 
-  setup [GOPATH](https://gist.github.com/vsouza/77e6b20520d07652ed7d) for your env file (.bashrc or .zshrc)
- ` git clone ` this repo into ` ~/golang/src `
- ` glide install ` in the local folder
- ` go get github.com/codegangsta/gin ` to install the reloading dev server

## Local Config
- rename `config.bak` to `config.json`
- add the proper keys - AWS_TOKEN is optional

## Local Run
- `gin` in the local folder then POST to [localhost:3000/upload](http://localhost:3000/upload) with a file to the `file` key
- will return either an error or a s3 url of saved file - the file is a uuid.file_extension

## Recommended Go Editor
- [Visual Studio Code](https://code.visualstudio.com/) with [Go Extension](https://github.com/Microsoft/vscode-go)

## TODO:
- test support for ENV VARS
- setup with docker
- setup proper package on github
- write tests?
