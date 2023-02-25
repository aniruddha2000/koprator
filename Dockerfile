FROM golang:alpine3.17 as builder
RUN mkdir "/src"
ADD . /src/
WORKDIR /src
RUN go build -ldflags "-s -w -X main.version=$(cat VERSION)" -o koprator
FROM alpine
COPY --from=builder /src/koprator /app/koprator
WORKDIR /app
ENTRYPOINT ["/app/koprator"]