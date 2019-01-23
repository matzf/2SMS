package main

import "net/http"

func GetMetricsInfoForMapping(mapping string, client *http.Client) []*MetricInfo {
	resp, err := LocalhostGet(mapping, client)
	if err != nil {
		return []*MetricInfo{}
	}
	// Parse response body to metric families
	metrics := DecodeResponseBody(resp)
	// Keep only name, type and help fields
	metricsInfo := make([]*MetricInfo, len(metrics))
	for i, metric := range metrics {
		metricsInfo[i] = &MetricInfo{*metric.Name, metric.Type.String(), *metric.Help}
	}
	return metricsInfo
}
