FROM golang:1.18-bullseye as build

RUN mkdir /app
WORKDIR /app

COPY ./internal ./internal
COPY ./cmd ./cmd
COPY ./configs ./configs
COPY go.mod ./go.mod
COPY go.sum ./go.sum

RUN go build -o /monitect-server cmd/server.go

# deploy image -- this is tiny! <100 mb for the whole image
FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=build /monitect-server /monitect-server
COPY --from=build /app/configs ./configs
EXPOSE 8080
ENTRYPOINT ["/monitect-server"]
CMD ["./configs/server.yaml"]
