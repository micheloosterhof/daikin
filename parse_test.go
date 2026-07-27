// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Tests for parsing the Daikin adapter's key=value response format,
// ABOUTME: including URL-encoded names and error responses.
package daikin

import "testing"

func TestParseResponse(t *testing.T) {
	kv, err := ParseResponse("ret=OK,pow=1,mode=3,adv=,stemp=23.0,name=%4f%66%66%69%63%65")
	if err != nil {
		t.Fatalf("ParseResponse returned error: %v", err)
	}
	want := map[string]string{
		"ret":   "OK",
		"pow":   "1",
		"mode":  "3",
		"adv":   "",
		"stemp": "23.0",
		"name":  "%4f%66%66%69%63%65",
	}
	for k, v := range want {
		if kv[k] != v {
			t.Errorf("kv[%q] = %q, want %q", k, kv[k], v)
		}
	}
}

func TestParseResponseNotOK(t *testing.T) {
	_, err := ParseResponse("ret=PARAM NG,adv=")
	if err == nil {
		t.Fatal("expected error for ret=PARAM NG")
	}
}

func TestParseResponseGarbage(t *testing.T) {
	_, err := ParseResponse("<html>forbidden</html>")
	if err == nil {
		t.Fatal("expected error for non key=value payload")
	}
}

func TestDecodeName(t *testing.T) {
	got, err := DecodeName("%4b%69%64%73%20%42%65%64%72%6f%6f%6d")
	if err != nil {
		t.Fatalf("DecodeName returned error: %v", err)
	}
	if got != "Kids Bedroom" {
		t.Errorf("DecodeName = %q, want %q", got, "Kids Bedroom")
	}
}
