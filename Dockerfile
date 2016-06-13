FROM golang:1.6

RUN \
  wget --no-check-certificate https://github.com/Masterminds/glide/releases/download/0.8.3/glide-0.8.3-linux-amd64.tar.gz && \
  tar xvf glide-0.8.3-linux-amd64.tar.gz && \
  mv linux-amd64/glide /usr/bin/ && \
  apt-get update && \
  apt-get install -y libmagic-dev
WORKDIR /go/src/github.com/bevanhunt/s3box
COPY . .
RUN glide install
RUN go build -o /go/bin/s3box .
CMD /go/bin/s3box
ENV PORT=8080
ENV GIN_MODE=release
EXPOSE 8080
