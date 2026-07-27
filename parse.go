// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Parses the Daikin wifi adapter's response format: comma-separated
// ABOUTME: key=value pairs with URL-encoded unit names.
package daikin

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseResponse parses a Daikin key=value response body into a map.
// It returns an error when the payload is not in that format or the
// adapter reported a non-OK ret code.
func ParseResponse(body string) (map[string]string, error) {
	kv := make(map[string]string)
	for pair := range strings.SplitSeq(strings.TrimSpace(body), ",") {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("malformed response segment %q", pair)
		}
		kv[k] = v
	}
	if ret := kv["ret"]; ret != "OK" {
		return nil, fmt.Errorf("adapter returned ret=%s", ret)
	}
	return kv, nil
}

// DecodeName decodes the URL-encoded unit name from basic_info.
func DecodeName(encoded string) (string, error) {
	return url.QueryUnescape(encoded)
}
