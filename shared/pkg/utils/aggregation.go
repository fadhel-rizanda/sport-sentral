package utils

import (
	sportv1 "microservice-golang/gen/sport/v1"
	"strings"
)

// ParseAggregationMethod converts a string (e.g. "SUM", "AVG") to sportv1.AggregationMethod enum.
func ParseAggregationMethod(s string) sportv1.AggregationMethod {
	switch strings.ToUpper(s) {
	case "SUM":
		return sportv1.AggregationMethod_AGGREGATION_METHOD_SUM
	case "AVG":
		return sportv1.AggregationMethod_AGGREGATION_METHOD_AVG
	case "MAX":
		return sportv1.AggregationMethod_AGGREGATION_METHOD_MAX
	case "MIN":
		return sportv1.AggregationMethod_AGGREGATION_METHOD_MIN
	default:
		return sportv1.AggregationMethod_AGGREGATION_METHOD_UNSPECIFIED
	}
}

// AggregationMethodToString converts a sportv1.AggregationMethod enum to its string representation (e.g. "SUM", "AVG").
func AggregationMethodToString(m sportv1.AggregationMethod) string {
	switch m {
	case sportv1.AggregationMethod_AGGREGATION_METHOD_SUM:
		return "SUM"
	case sportv1.AggregationMethod_AGGREGATION_METHOD_AVG:
		return "AVG"
	case sportv1.AggregationMethod_AGGREGATION_METHOD_MAX:
		return "MAX"
	case sportv1.AggregationMethod_AGGREGATION_METHOD_MIN:
		return "MIN"
	default:
		return "AVG" // Fallback default
	}
}
