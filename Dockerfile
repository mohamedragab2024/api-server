FROM golang:1.16
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go get github.com/go-redis/redis/v8
RUN go build -o main .
CMD ["/app/main"]