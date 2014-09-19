package main

import (
	"fmt"
	"regexp"
)

type rule struct {
	src string
	dst string
}

/*
   ipaddr = \d+\.\d+\.\d+\.\d+
   port   = \d+
   portsep = :
   fulladdress = <ipaddr> <portsep> <port>
   portonly    = <port>
*/

var reRule *regexp.Regexp

func init() {
	ipaddr := `\d+\.\d+\.\d+\.\d+`
	port := `\d+`
	portSep := `:`
	fullAddress := ipaddr + portSep + port

	addrPattern := fmt.Sprintf("(?:%s|%s)", fullAddress, port)
	rulePattern := fmt.Sprintf(`^\s*(%s)\s*->\s*(%s)\s*`, addrPattern, addrPattern)
	reRule = regexp.MustCompile(rulePattern)
}

func parseRule(r string) (*rule, error) {
	matches := reRule.FindAllStringSubmatch(r, -1)
	if matches == nil {
		return nil, fmt.Errorf("Failed to match '%s'", r)
	}

	return &rule{matches[0][1], matches[0][2]}, nil
}
