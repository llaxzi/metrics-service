package apperrors

import "errors"

var (
	ErrWrongMetricValue  = errors.New("wrong metric value")
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrMetricNotExist    = errors.New("metric doesn't exist")
	ErrServer            = errors.New("server error")
	ErrPgConnExc         = errors.New("pg connection Exception")
	ErrPingMemory        = errors.New("trying to ping memory storage")
)

func IsAppError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrWrongMetricValue) ||
		errors.Is(err, ErrInvalidMetricType) ||
		errors.Is(err, ErrMetricNotExist) ||
		errors.Is(err, ErrServer) ||
		errors.Is(err, ErrPgConnExc) ||
		errors.Is(err, ErrPingMemory)
}
