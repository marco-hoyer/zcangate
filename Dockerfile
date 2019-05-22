FROM arm64v8/alpine:3.9

ADD zcangate /usr/bin/zcangate

RUN apk update && apk add ca-certificates

ENTRYPOINT ["zcangate"]