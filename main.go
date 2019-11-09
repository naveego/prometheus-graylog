package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/naveego/prometheus-graylog/internal/log"
	"github.com/golang/snappy"
	"github.com/gogo/protobuf/proto"
	"github.com/rs/xid"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/Graylog2/go-gelf/gelf"
)

const MicroserviceName = "prometheus-graylog"

var gelfWriter *gelf.Writer

func main() {
	log.Info("Starting prometheus graylog write endpoint")
	var err error

	gelfWriter, err = gelf.NewWriter("")
	if err != nil {
		log.WithError(err).Fatal("Could not create gelf writer")
	}

	r := mux.NewRouter()
	r.HandleFunc("/receive", receiveHandler)

	go func() {
		log.Fatal(http.ListenAndServe(":8081", r))
	}()

	log.Infof("Listening for requests on :%d", 8081)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Info("Received CTRL-C, shutting down daemon")
}

type metricLog struct {
	Labels map[string]string `json:"labels"`
	Value float64 `json:"value"`
}

func receiveHandler(w http.ResponseWriter, r *http.Request) {
	correlationID := xid.New().String()
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("couldn't read body")
		return
	}
	defer r.Body.Close()

	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithError(err).Error("couldn't decompress body")
		return
	}

	var req prompb.WriteRequest
	if err := proto.Unmarshal(reqBuf, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithError(err).Error("couldn't unmarshal body")
		return
	}

	// Loop over the time series and write them to gelf
	for _, ts := range req.Timeseries {
		labels := make(map[string]string, len(ts.Labels))

		for _, l := range ts.Labels {
			labels[string(model.LabelName(l.Name))] = string(model.LabelValue(l.Value))
		}

		for _, s := range ts.Samples {
			epoch := time.Unix(s.Timestamp/1000, 0).UTC().Unix()
			host, _ := os.Hostname()

			extra := map[string]interface{}{
				"microservice": MicroserviceName,
				"correlation_id": correlationID,
				"metric": metricLog{
					Labels: labels,
					Value:  s.Value,
				},
			}

			msg := &gelf.Message{
				Version:  "v1",
				Host:     host,
				Short:    "",
				Full:     "",
				TimeUnix: float64(epoch),
				Level:    6,
				Facility: MicroserviceName,
				Extra:    extra,
			}

			err = gelfWriter.WriteMessage(msg)
			if err != nil {
				log.WithError(err).Warn("Could not send metric message to graylog")
			}
		}
	}
}