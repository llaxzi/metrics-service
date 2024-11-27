package collector

import (
	"math/rand"
	"runtime"
)

type MetricsCollector interface {
	Collect() map[string]interface{}
}

type metricsCollector struct {
}

func NewMetricsCollector() MetricsCollector {
	return &metricsCollector{}
}

func (m *metricsCollector) Collect() map[string]interface{} {
	var memStats runtime.MemStats
	metricsMap := make(map[string]interface{})

	runtime.ReadMemStats(&memStats)
	metricsMap["Alloc"] = memStats.Alloc
	metricsMap["BuckHashSys"] = memStats.BuckHashSys
	metricsMap["Frees"] = memStats.Frees
	metricsMap["GCCPUFraction"] = memStats.GCCPUFraction
	metricsMap["HeapAlloc"] = memStats.HeapAlloc
	metricsMap["HeapIdle"] = memStats.HeapIdle
	metricsMap["HeapInuse"] = memStats.HeapInuse
	metricsMap["HeapObjects"] = memStats.HeapObjects
	metricsMap["HeapReleased"] = memStats.HeapReleased
	metricsMap["HeapSys"] = memStats.HeapSys
	metricsMap["LastGC"] = memStats.LastGC
	metricsMap["Lookups"] = memStats.Lookups
	metricsMap["MCacheInuse"] = memStats.MCacheInuse
	metricsMap["MCacheSys"] = memStats.MCacheSys
	metricsMap["MSpanInuse"] = memStats.MSpanInuse
	metricsMap["MSpanSys"] = memStats.MSpanSys
	metricsMap["Mallocs"] = memStats.Mallocs
	metricsMap["NextGC"] = memStats.NextGC
	metricsMap["NumForcedGC"] = memStats.NumForcedGC
	metricsMap["NumGC"] = memStats.NumGC
	metricsMap["OtherSys"] = memStats.OtherSys
	metricsMap["PauseTotalNs"] = memStats.PauseTotalNs
	metricsMap["StackInuse"] = memStats.StackInuse
	metricsMap["StackSys"] = memStats.StackSys
	metricsMap["Sys"] = memStats.Sys
	metricsMap["TotalAlloc"] = memStats.TotalAlloc

	metricsMap["RandomValue"] = rand.Float64()

	return metricsMap
}
