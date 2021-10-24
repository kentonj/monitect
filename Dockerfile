FROM golang:1.17

WORKDIR /go/monitect

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY cmd/server cmd/server
COPY src ./src
COPY conf/server.yaml ./conf/server.yaml

RUN go build -o server cmd/server/server.go
RUN chmod +x ./server
CMD ["./server"]
