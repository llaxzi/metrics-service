package storage

import (
	"sync"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestDiskStorage(t *testing.T) {
	fName := `metrics.json`

	metricsSt := &metricsStorage{sync.RWMutex{}, sync.RWMutex{}, make(map[string]float64), make(map[string]int64), nil}
	metricsSt.setGauge("nameG", 10)
	metricsSt.setCounter("nameC", 2)

	diskW, _ := NewDiskWriter(fName)
	err := diskW.Save(metricsSt.getMetricsJSON())
	if err != nil {
		t.Error(err)
	}
	saveResult := metricsSt.getMetricsJSON()

	metricsSt = &metricsStorage{sync.RWMutex{}, sync.RWMutex{}, make(map[string]float64), make(map[string]int64), nil}
	diskR, _ := NewDiskReader(fName)
	err = diskR.Load(metricsSt)
	if err != nil {
		t.Error(err)
	}
	loadResult := metricsSt.getMetricsJSON()
	diskR.Close()

	assert.Equal(t, saveResult, loadResult)
}
