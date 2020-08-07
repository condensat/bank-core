FROM golang:1.14-alpine

RUN apk add --no-cache git

ARG GITLAB_BUILD_TOKEN=

RUN git config --global credential.helper store
RUN git config --global url."https://build:${GITLAB_BUILD_TOKEN}@code.condensat.tech/".insteadOf https://code.condensat.tech/
RUN go env -w GOPRIVATE=code.condensat.tech/*

ENV CGO_ENABLED=0

WORKDIR /src
ADD go.mod go.sum /src/
RUN go mod download
