package main

import "testing"

func TestParseRule(t *testing.T) {
	checkParseResult(t, "11212 -> 11211", "11212", "11211")
	checkParseResult(t, "192.168.0.1:11212 -> 192.168.0.2:11211", "192.168.0.1:11212", "192.168.0.2:11211")
}

func checkParseResult(t *testing.T, pat, src, dst string) {
	var err error
	//var r proxyRunner
	if _, err = parseRule(pat); err != nil {
		t.Errorf("Failed to parse: %s", err)
		return
	}
/*

	if r.src != src {
		t.Errorf("expected r.src = %s, got %s", src, r.src)
		return
	}
	if r.dst != dst {
		t.Errorf("expected r.dst = %s, got %s", dst, r.dst)
		return
	}
*/
}