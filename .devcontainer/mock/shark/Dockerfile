FROM golang:1.23.2-alpine AS build

WORKDIR /go/src/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/app main.go

FROM alpine:3.20.3

LABEL com.github.Br0ce.opera.tool.name="get_shark_warning"
LABEL com.github.Br0ce.opera.tool.port="8080"
LABEL com.github.Br0ce.opera.tool.path="/"

RUN apk --no-cache add ca-certificates
COPY --from=build /go/src/bin/app ./home

WORKDIR /home

RUN addgroup -S appgroup --gid 1001 && adduser -S --no-create-home appuser -G appgroup --uid 1001
RUN chown appuser:appgroup ./app
USER appuser

ENTRYPOINT ["./app"]