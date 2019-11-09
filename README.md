# Prometheus to Graylog Storage Adapter
The [Prometheus] to [Graylog] storage adapter is an implementation of the [Prometheus] 
[remote_write](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write)
api.  It is used to maintain a long term history of [Prometheus] metrics for later analysis.

This code was inspired by the https://github.com/Telefonica/prometheus-kafka-adapter project.

[Prometheus]: https://prometheus.io
[Graylog]: https://www.graylog.org 