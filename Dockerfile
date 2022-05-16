FROM golang:1.18-alpine
WORKDIR /app
COPY . .
EXPOSE 8080
RUN go build -o bitcask-kv
ENTRYPOINT [ "./bitcask-kv", "-server=true", "-port=8080" ]
