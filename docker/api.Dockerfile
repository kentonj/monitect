FROM golang:1.18-bullseye as build

RUN mkdir /build
WORKDIR /build

COPY ./internal ./internal
COPY ./cmd ./cmd
COPY go.mod ./go.mod
COPY go.sum ./go.sum

RUN go mod download
# build binary to /go/bin
RUN go build -o /go/bin/api cmd/api.go

# deploy image -- this is tiny! <100 mb for the whole image
FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/api /
COPY ./configs /configs
EXPOSE 8080
ENTRYPOINT ["/api"]
