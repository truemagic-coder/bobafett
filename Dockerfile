FROM golang:1.6

RUN \
  wget --no-check-certificate https://github.com/Masterminds/glide/releases/download/v0.11.0/glide-v0.11.0-linux-amd64.tar.gz && \
  tar xvf glide-v0.11.0-linux-amd64.tar.gz && \
  mv linux-amd64/glide /usr/bin/ && \
  apt-get update && \
  apt-get install -y libmagic-dev
WORKDIR /go/src/github.com/bevanhunt/bobafett
COPY . .
RUN glide install
RUN go build -o /go/bin/bobafett .
CMD /go/bin/bobafett
ENV PORT=8080
ENV GIN_MODE=release
EXPOSE 8080
