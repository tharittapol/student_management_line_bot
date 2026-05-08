# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/student-management-line-bot .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=build /out/student-management-line-bot /app/student-management-line-bot
COPY db/schema.sql /app/db/schema.sql

ENV TZ=Asia/Bangkok
ENV PORT=8080

EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/app/student-management-line-bot"]
