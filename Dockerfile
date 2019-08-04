FROM node:12.7-alpine AS frontend-builder
COPY . /app/src
WORKDIR /app/src/frontend
RUN npm i
RUN npm run build
RUN cp -r dist/* ../public/

FROM golang:1.12 AS build-env
ENV CGO_ENABLED=0
COPY . /app/src
WORKDIR /app/src
RUN go build -a -ldflags '-extldflags "-static"' -mod=vendor ipprovider.go

FROM alpine:3.9
LABEL maintainer="me@iskywind.com"
RUN apk add --update iptables && rm -rf /vat/cache/apk/*

COPY --from=build-env /app/src/ipprovider /app/
COPY --from=frontend-builder /app/src/public /app/public
WORKDIR /app

ENTRYPOINT ["./ipprovider"]
