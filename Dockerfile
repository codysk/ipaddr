FROM alpine:3.9
LABEL maintainer="me@iskywind.com"

COPY ipprovider /app/
WORKDIR /app

ENTRYPOINT ["./ipprovider"]
