FROM golang
ADD . /go/src/zcangate
WORKDIR /go/src/zcangate
RUN env GOOS=linux GOARCH=arm GOARM=7 go build

FROM arm64v8/alpine:3.13
COPY --from=0 /go/src/zcangate/zcangate /usr/bin/zcangate

EXPOSE 8080
ENTRYPOINT ["zcangate"]
