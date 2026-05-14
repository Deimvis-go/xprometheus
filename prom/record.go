package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
)

type RecordControl interface {
	GetLabels() prometheus.Labels
	AddLabels(ls prometheus.Labels)
	SetLabels(ls prometheus.Labels)
}

type RecordFn func(rc RecordControl)

func NewRecordControl() RecordControl {
	return &recordControl{
		ls: make(prometheus.Labels),
	}
}

type recordControl struct {
	ls prometheus.Labels
}

func (rc *recordControl) GetLabels() prometheus.Labels {
	return rc.ls
}

func (rc *recordControl) SetLabels(ls prometheus.Labels) {
	rc.ls = ls
}

func (rc *recordControl) AddLabels(ls prometheus.Labels) {
	for k, v := range ls {
		rc.ls[k] = v
	}
}

func Record(fns ...RecordFn) {
	xmust.True(len(fns) == 1, "TODO: support middleware chaining")
	fn := fns[0]
	rc := NewRecordControl()
	fn(rc)
}
