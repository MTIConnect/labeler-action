FROM golang:1.13.6-alpine

WORKDIR /go/src/app
COPY . .
RUN go build -mod=vendor
ENTRYPOINT ["/go/src/app/labeler-action"]
