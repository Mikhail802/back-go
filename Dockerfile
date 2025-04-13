FROM golang:alpine AS builder

WORKDIR /

RUN apk add --no-cache git

ADD go.mod ./
ADD go.sum ./
RUN go mod download

COPY . .

RUN go get gorm.io/datatypes
RUN go build -o main main.go

FROM alpine

RUN apk add curl

WORKDIR /

COPY --from=builder /main /main
COPY app.env /app.env

EXPOSE 8000

CMD ["/main", "--host", "0.0.0.0", "--port", "8000"]