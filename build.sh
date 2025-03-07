#!/bin/sh

APP=dify-apps-exporter

mk() {
    go build -trimpath -o $APP-$GOOS-$GOARCH .
}

GOOS=linux GOARCH=amd64 mk
GOOS=linux GOARCH=arm64 mk
GOOS=darwin GOARCH=amd64 mk
GOOS=darwin GOARCH=arm64 mk
GOOS=windows GOARCH=amd64 mk
