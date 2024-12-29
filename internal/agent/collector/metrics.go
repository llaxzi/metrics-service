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
	metricsMap["Alloc"] = float64(memStats.Alloc)
	metricsMap["BuckHashSys"] = float64(memStats.BuckHashSys)
	metricsMap["Frees"] = float64(memStats.Frees)
	metricsMap["GCCPUFraction"] = memStats.GCCPUFraction
	metricsMap["HeapAlloc"] = float64(memStats.HeapAlloc)
	metricsMap["HeapIdle"] = float64(memStats.HeapIdle)
	metricsMap["HeapInuse"] = float64(memStats.HeapInuse)
	metricsMap["HeapObjects"] = float64(memStats.HeapObjects)
	metricsMap["HeapReleased"] = float64(memStats.HeapReleased)
	metricsMap["HeapSys"] = float64(memStats.HeapSys)
	metricsMap["LastGC"] = float64(memStats.LastGC)
	metricsMap["Lookups"] = float64(memStats.Lookups)
	metricsMap["MCacheInuse"] = float64(memStats.MCacheInuse)
	metricsMap["MCacheSys"] = float64(memStats.MCacheSys)
	metricsMap["MSpanInuse"] = float64(memStats.MSpanInuse)
	metricsMap["MSpanSys"] = float64(memStats.MSpanSys)
	metricsMap["Mallocs"] = float64(memStats.Mallocs)
	metricsMap["NextGC"] = float64(memStats.NextGC)
	metricsMap["NumForcedGC"] = float64(memStats.NumForcedGC)
	metricsMap["NumGC"] = float64(memStats.NumGC)
	metricsMap["OtherSys"] = float64(memStats.OtherSys)
	metricsMap["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	metricsMap["StackInuse"] = float64(memStats.StackInuse)
	metricsMap["StackSys"] = float64(memStats.StackSys)
	metricsMap["Sys"] = float64(memStats.Sys)
	metricsMap["TotalAlloc"] = float64(memStats.TotalAlloc)
	metricsMap["GCSys"] = float64(memStats.GCSys)

	metricsMap["RandomValue"] = rand.Float64()

	return metricsMap
}
