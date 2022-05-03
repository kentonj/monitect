FROM golang:1.17

WORKDIR /go/monitect

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY cmd/server cmd/server
COPY internal ./internal
COPY configs/server.yaml ./configs/server.yaml

RUN go build -o server cmd/server/main.go
RUN chmod +x ./server
CMD ["./server"]
