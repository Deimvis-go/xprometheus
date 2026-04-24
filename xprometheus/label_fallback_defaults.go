package xprometheus

const (
	// LabelNoValue is a placeholder for an empty value.
	// "No value" is better than label absence for ease of aggregation,
	// since metrics query engines may ignore records where a label is
	// absent on an aggregating operation.
	LabelNoValue = "-"
	// LabelUnknown indicates that a value MAY exist in reality,
	// but is unavailable to extract at the moment the metric is recorded.
	// For example, a response is received, but the code interface doesn't
	// allow access to the response status code, so set unknown instead of
	// no-value to indicate the value exists but is unobservable.
	LabelUnknown = "<unknown>"
	// LabelHighCardinality is used when a reasonable limit
	// for label cardinality has been reached.
	LabelHighCardinality = "<high_cardinality>"
)
