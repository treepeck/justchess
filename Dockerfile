FROM golang:1.22.5 AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api main/main.go

FROM alpine:latest
EXPOSE 3502
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates
COPY .env .
COPY ./pkg/db/schema.sql .
COPY --from=build /bin/api /bin/
ENTRYPOINT ["/bin/api"]