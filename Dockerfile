FROM golang:1.14-buster AS builder
COPY . /go/src/gitlab.inovex.io/proj-kosmos/kosmos-mediator
WORKDIR /go/src/gitlab.inovex.io/proj-kosmos/kosmos-mediator
RUN go build -ldflags "-linkmode external -extldflags -static" -o /usr/local/bin/mediator

FROM gcr.io/distroless/static-debian10:latest
COPY --from=builder /usr/local/bin/mediator /usr/local/bin/mediator
USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/mediator"]