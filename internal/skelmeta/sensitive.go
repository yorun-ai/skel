// Package skelmeta contains internal conventions shared across Skel analysis
// and code generation.
package skelmeta

import "go.yorun.ai/skel/internal/util/nameutil"

const SensitiveMarkerMethodName = "SkelSensitive"

// SensitiveMarkerFieldName returns the Skel field name that would collide with
// the generated marker method after Go name conversion.
func SensitiveMarkerFieldName() string {
	return nameutil.ToLowerCamel(SensitiveMarkerMethodName)
}
