FROM golang:alpine

WORKDIR /go/src/github.com/HenrySlawniak/jmaas
COPY . .

RUN go version

RUN apk add --update git bash
RUN bash -c "go build -ldflags '-w -X main.buildTime=$(date +'%b-%d-%Y-%H:%M:%S') -X main.commit=$(git describe --always --dirty=*)' -o jmaas ."

CMD ["./jmaas"]
