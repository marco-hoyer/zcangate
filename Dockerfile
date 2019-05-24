FROM golang
ADD . /go/src/zcangate
WORKDIR /go/src/zcangate
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN GOOS=linux GOARCH=arm GOARM=6 go build

FROM arm64v8/alpine:3.9
COPY --from=0 /go/src/zcangate /usr/bin/zcangate
RUN apk update && apk --no-cache add ca-certificates

EXPOSE 8080
ENTRYPOINT ["zcangate"]
