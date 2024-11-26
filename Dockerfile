FROM golang:1.22.5

WORKDIR /usr/src/app

COPY . .
RUN go mod tidy 