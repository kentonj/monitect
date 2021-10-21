FROM golang:1.17-alpine

WORKDIR /go/monitect

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY cmd/server cmd/server
COPY src ./src
COPY conf/server.yaml ./conf/server.yaml

RUN go build -o monitect cmd/server/main.go
RUN chmod +x ./monitect
CMD ["./monitect"]
