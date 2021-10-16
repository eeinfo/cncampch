# syntax=docker/dockerfile:1
FROM 1.17-alpine, 1-alpine as builder
##
## Build
##
WORKDIR /workspace

# Copy the go source
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY *.go .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go
 
##
## Deploy
##
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /http_server /http_server

EXPOSE 8010

USER nonroot:nonroot

ENTRYPOINT ["/http_server"]

