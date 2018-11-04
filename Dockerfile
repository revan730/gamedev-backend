FROM golang:latest AS BUILDER

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
ADD . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -mod=vendor -o gameback

FROM alpine:3.8

COPY --from=builder /go/src/app/gameback /bin/gameback

EXPOSE 8080
ENTRYPOINT ["/bin/gameback"]
CMD ["start"]
