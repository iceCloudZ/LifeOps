# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/lifeops-api .

FROM alpine:3.22
RUN adduser -D -H -s /sbin/nologin lifeops
USER lifeops
EXPOSE 8080
ENV LIFEOPS_ADDR=:8080
ENV LIFEOPS_DB_PATH=/data/lifeops.db
ENTRYPOINT ["/lifeops-api"]
COPY --from=build /out/lifeops-api /lifeops-api
