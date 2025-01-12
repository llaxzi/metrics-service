package apperrors

import "errors"

var (
	ErrWrongMetricValue  = errors.New("wrong metric value")
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrMetricNotExist    = errors.New("metric doesn't exist")
	ErrServer            = errors.New("server error")
	ErrPgConnExc         = errors.New("pg connection Exception")
)
