FROM golang:1.14-buster AS builder
COPY . /go/src/github.kosmos.io/proj-kosmos/kosmos-mediator
WORKDIR /go/src/github.kosmos.io/proj-kosmos/kosmos-mediator
RUN go build -o /usr/local/bin/mediator

FROM gcr.io/distroless/base-debian10:latest
COPY --from=builder /usr/local/bin/mediator /usr/local/bin/mediator
USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/mediator"]
