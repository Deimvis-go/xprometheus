package prom

import (
	"net/http"
	"strconv"

	"github.com/Deimvis/go-ext/go1.25/xfallback/xfb"
	"github.com/Deimvis/go-ext/go1.25/xhttp"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/prometheus/client_golang/prometheus"
)

type RoundTripWrapOption func(*roundTripWrapCfg)

func NewRoundTripWrap(ig IntervalMetricGroup, opts ...RoundTripWrapOption) xhttp.RoundTripWrapFn {
	cfg := defautRoundTripWrapConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(fn xhttp.RoundTripFn) xhttp.RoundTripFn {
		return func(req *http.Request) (*http.Response, error) {
			var resp *http.Response
			var err error

			ig.Record(cfg.newStartLabels(&cfg, req), func(rc RecordControl) {
				resp, err = fn(req)
				rc.SetLabels(cfg.newFinishLabels(&cfg, req, resp, err))
			})

			return resp, err
		}

	}
}

func WithStartLabelsFn(fn StartLabelsFn) RoundTripWrapOption {
	return func(cfg *roundTripWrapCfg) {
		cfg.newStartLabels = fn
	}
}

func WithFinishLabelsFn(fn FinishLabelsFn) RoundTripWrapOption {
	return func(cfg *roundTripWrapCfg) {
		cfg.newFinishLabels = fn
	}
}

func WithNormalizer(norm xhttp.PathNormalizer) RoundTripWrapOption {
	return func(cfg *roundTripWrapCfg) {
		cfg.norm.SetValue(norm)
	}
}

type RoundTripWrapConfig interface {
	Normalizer() (xhttp.PathNormalizer, bool)
}

type StartLabelsFn = func(RoundTripWrapConfig, *http.Request) prometheus.Labels
type FinishLabelsFn = func(RoundTripWrapConfig, *http.Request, *http.Response, error) prometheus.Labels

type roundTripWrapCfg struct {
	norm xoptional.T[xhttp.PathNormalizer]
	// TODO: figure out proper interface to do not calculate labels twice
	newStartLabels  StartLabelsFn
	newFinishLabels FinishLabelsFn
}

var _ RoundTripWrapConfig = (*roundTripWrapCfg)(nil)

func (cfg *roundTripWrapCfg) Normalizer() (xhttp.PathNormalizer, bool) {
	return xoptional.ValueWithOk(cfg.norm)
}

var (
	// start labels: host, method, path
	// finish labels: host, method, path, code
	defautRoundTripWrapConfig = roundTripWrapCfg{
		norm:            xoptional.T[xhttp.PathNormalizer]{},
		newStartLabels:  DefaultStartLabelsFn,
		newFinishLabels: DefaultFinishLabelsFn,
	}
	DefaultStartLabelsFn = func(this RoundTripWrapConfig, req *http.Request) prometheus.Labels {
		ls := prometheus.Labels{
			"host":   req.URL.Hostname(),
			"method": req.Method,
		}

		path := LabelUnknown
		if norm, ok := this.Normalizer(); ok {
			path = xfb.OnNilv(norm(req), LabelUnknown)
		}
		ls["path"] = path

		return ls
	}
	DefaultFinishLabelsFn = func(this RoundTripWrapConfig, req *http.Request, resp *http.Response, err error) prometheus.Labels {
		ls := prometheus.Labels{
			"host":   req.URL.Hostname(),
			"method": req.Method,
		}

		path := LabelUnknown
		if norm, ok := this.Normalizer(); ok {
			path = xfb.OnNilv(norm(req), LabelUnknown)
		}
		ls["path"] = path

		code := LabelNoValue
		if err == nil {
			code = strconv.Itoa(resp.StatusCode)
		}
		ls["code"] = code
		return ls
	}
)
