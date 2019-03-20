FROM golang:1.12 as build

ENV GOPROXY https://go.likeli.top
ENV GO111MODULE on

WORKDIR /go/cache

COPY go.mod .
COPY go.sum .
RUN go mod download

WORKDIR /go/release

COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix cgo -o main main.go

FROM scratch as prod

COPY --from=build /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/release/main /
COPY --from=build /go/release/config.yaml /

CMD ["/main"]
