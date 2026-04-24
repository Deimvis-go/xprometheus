package xprometheus

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type RoundTripWrapOption func(*roundTripWrapCfg)

func NewRoundTripWrap(ig IntervalMetricGroup, opts ...RoundTripWrapOption) RoundTripWrapFn {
	cfg := defautRoundTripWrapConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(fn RoundTripFn) RoundTripFn {
		return func(req *http.Request) (*http.Response, error) {
			var resp *http.Response
			var err error

			ig.Record(cfg.newStartLabels(cfg, req), func(rc RecordControl) {
				resp, err = fn(req)
				rc.SetLabels(cfg.newFinishLabels(cfg, req, resp, err))
			})

			return resp, err
		}
	}
}

func WithNormalizer(norm PathNormalizer) RoundTripWrapOption {
	return func(cfg *roundTripWrapCfg) {
		cfg.norm = norm
	}
}

type roundTripWrapCfg struct {
	norm            PathNormalizer
	newStartLabels  func(roundTripWrapCfg, *http.Request) prometheus.Labels
	newFinishLabels func(roundTripWrapCfg, *http.Request, *http.Response, error) prometheus.Labels
}

var defautRoundTripWrapConfig = roundTripWrapCfg{
	newStartLabels: func(this roundTripWrapCfg, req *http.Request) prometheus.Labels {
		ls := prometheus.Labels{
			"host":   req.URL.Hostname(),
			"method": req.Method,
		}

		path := LabelUnknown
		if this.norm != nil {
			path = ptrDerefOr(this.norm(req), LabelUnknown)
		}
		ls["path"] = path

		return ls
	},
	newFinishLabels: func(this roundTripWrapCfg, req *http.Request, resp *http.Response, err error) prometheus.Labels {
		ls := prometheus.Labels{
			"host":   req.URL.Hostname(),
			"method": req.Method,
		}

		path := LabelUnknown
		if this.norm != nil {
			path = ptrDerefOr(this.norm(req), LabelUnknown)
		}
		ls["path"] = path

		code := LabelNoValue
		if err == nil {
			code = strconv.Itoa(resp.StatusCode)
		}
		ls["code"] = code
		return ls
	},
}
