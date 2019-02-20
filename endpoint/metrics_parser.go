package main

import (
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"github.com/netsec-ethz/2SMS/common"
	"io"
	"log"
	"mime"
	"net/http"
)

// Taken from: prom2json.ParseResponse
func DecodeResponseBody(resp *http.Response) []*common.MetricFamily {
	metrics := []*common.MetricFamily{}
	mediaType, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err == nil && mediaType == "application/vnd.google.protobuf" &&
		params["encoding"] == "delimited" &&
		params["proto"] == "io.prometheus.client.MetricFamily" {
		for {
			mf := &common.MetricFamily{}
			if _, err = pbutil.ReadDelimited(resp.Body, mf); err != nil {
				if err == io.EOF {
					break
				}
				//return fmt.Errorf("reading metric family protocol buffer failed: %v", err)
			}
			metrics = append(metrics, mf)
		}
	} else {
	// TODO: removing these prometheus dependencies seems quite hard so for the moment just don't support the fallback
		log.Printf("MetricsParser: %s Media Type is not supported, decoding failed.", mediaType)
	//	// We could do further content-type checks here, but the
	//	// fallback for now will anyway be the text format
	//	// version 0.0.4, so just go for it and see if it works.
	//	var parser expfmt.TextParser
	//	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	//	if err != nil {
	//		//return fmt.Errorf("reading text format failed: %v", err)
	//	}
	//	for _, mf := range metricFamilies {
	//		metrics = append(metrics, mf)
	//	}
	}
	return metrics
}
