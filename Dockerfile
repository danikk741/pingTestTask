FROM golang:latest

LABEL maintainer="Daniyar Bexoltanov <beksoltanov98@gmail.com>"

WORKDIR /app

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]