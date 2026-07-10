package timeseries

import "chronocone/backend/domain"

type TimeSeriesDocType struct {
	domain.Contained
}

type TimeSeriesRowType struct {
	domain.Contained
}

type TimeSeriesDocLineage struct {
	domain.Contained
}
