// Package collector предоставляет функциональность для сбора системных метрик.
package collector

import (
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// IMetricsCollector определяет интерфейс сборщика метрик.

type IMetricsCollector interface {
	// Collect возвращает мапу с метриками.
	Collect() map[string]interface{}
}

// metricsCollector представляет реализацию IMetricsCollector.
type metricsCollector struct {
}

// NewMetricsCollector создает и возвращает новый экземпляр IMetricsCollector.
func NewMetricsCollector() IMetricsCollector {
	return &metricsCollector{}
}

// Collect собирает различные системные метрики, включая:
// - Статистику работы runtime
// - Использование памяти
// - Загрузку процессора
// - Случайное значение (RandomValue) для тестирования;
// Возвращает карту, содержащую названия метрик и их значения.
func (m *metricsCollector) Collect() map[string]interface{} {
	//runtime
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

	// gopsutil
	v, _ := mem.VirtualMemory()
	metricsMap["TotalMemory"] = float64(v.Total)
	metricsMap["FreeMemory"] = float64(v.Free)
	percentages, _ := cpu.Percent(1*time.Second, true)
	for i, percentage := range percentages {
		metricsMap["CPUutilization"+strconv.Itoa(i+1)] = percentage
	}

	return metricsMap
}
