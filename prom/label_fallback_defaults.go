package prom

const (
	// No value is a placeholder for empty value.
	// No value is better than label absence for ease of aggregation,
	// since metrics query engine may ignore records
	// where label is absent on aggregating operation.
	LabelNoValue = "-"
	// Unknown is used to indicate that value MAY exist in reality,
	// but is unavailable to extract on the moment metric is recorded.
	// For example, response is received, but
	// code interface doesn't allow you to access response status code
	// and instead of settings no value you would rather set unknown value
	// to indicate that response status code may exist, but you don't know it.
	LabelUnknown = "<unknown>"
	// LabelHighCardinality is used when a reasonable limit
	// for label cardinality has been reached.
	LabelHighCardinality = "<high_cardinality>"
)
