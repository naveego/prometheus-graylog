FROM golang:1.13.3-alpine3.10 as build

WORKDIR /src/prometheus-graylog-adapter
ADD . /src/prometheus-graylog-adapter

RUN go build -o /prometheus-graylog-adapter

FROM alpine:3.10

COPY --from=build /prometheus-graylog-adapter /

CMD /prometheus-graylog-adapter