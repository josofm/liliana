FROM golang:1.22 as devimage

EXPOSE 80
WORKDIR /app
COPY go.mod /app
RUN go mod download
RUN go mod tidy
COPY . /app

ENTRYPOINT ["/app/liliana"]
