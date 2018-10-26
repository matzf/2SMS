package main

import (
	"net/http"
	"github.com/prometheus/client_model/go"
	"bytes"
	"github.com/prometheus/common/expfmt"
	"mime"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"io"
	dto "github.com/prometheus/client_model/go"
)

// Taken from: prom2json.ParseResponse
func DecodeResponseBody(resp *http.Response) []*io_prometheus_client.MetricFamily{
	metrics := []*io_prometheus_client.MetricFamily{}
	mediatype, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err == nil && mediatype == "application/vnd.google.protobuf" &&
		params["encoding"] == "delimited" &&
		params["proto"] == "io.prometheus.client.MetricFamily" {
		for {
			mf := &dto.MetricFamily{}
			if _, err = pbutil.ReadDelimited(resp.Body, mf); err != nil {
				if err == io.EOF {
					break
				}
				//return fmt.Errorf("reading metric family protocol buffer failed: %v", err)
			}
			metrics = append(metrics, mf)
		}
	} else {
		// We could do further content-type checks here, but the
		// fallback for now will anyway be the text format
		// version 0.0.4, so just go for it and see if it works.
		var parser expfmt.TextParser
		metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
		if err != nil {
			//return fmt.Errorf("reading text format failed: %v", err)
		}
		for _, mf := range metricFamilies {
			metrics = append(metrics, mf)
		}
	}
	return metrics
}

// Taken from: client_golang/prometheus/promhttp/http.HandleFor line 142
func EncodeMetrics(metrics []*io_prometheus_client.MetricFamily, format expfmt.Format) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := expfmt.NewEncoder(buf, format)
	for _, mf := range metrics {
		if err := enc.Encode(mf); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}