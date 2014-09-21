package main

import (
	"fmt"
	"github.com/lestrrat/go-lex"
	"regexp"
	"unicode"
)

type rule struct {
	src string
	dst string
}

type ruleParser struct {
	*lex.ItemConsume
}

const (
	ItemPortNumber = iota + 1000
	ItemHostname
	ItemWhitespace
	ItemArrow
	ItemColon
)

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

func lexStart(l lex.Lexer, ctx interface{}) lex.LexFn {
	return lexSource
}

func scanWhitespace(l lex.Lexer, ctx interface{}) bool {
	if n := l.Peek(); !unicode.IsSpace(n) {
		return false
	}

	l.Next()
	for loop := true; loop; {
		n := l.Peek()
		switch {
		case unicode.IsSpace(n):
			l.Next()
		default:
			loop = false
		}
	}
	return true
}

func lexArrow(l lex.Lexer, ctx interface{}) lex.LexFn {
	if scanWhitespace(l, ctx) {
		l.Emit(ItemWhitespace)
	}

	if !l.AcceptString("->") {
		l.EmitErrorf("lex: Expected '->'")
	}
	l.Emit(ItemArrow)
	return lexDestination
}

func lexDestination(l lex.Lexer, ctx interface{}) lex.LexFn {
	return lexHostnameOrPort(l, ctx, nil)
}

func lexSource(l lex.Lexer, ctx interface{}) lex.LexFn {
	return lexHostnameOrPort(l, ctx, lexArrow)
}

func lexHostnameOrPort(l lex.Lexer, ctx interface{}, next lex.LexFn) lex.LexFn {
	if scanWhitespace(l, ctx) {
		l.Emit(ItemWhitespace)
	}

	// ipaddr, hostname followed by a colon and port number,
	// or just port number
	if lex.AcceptRun(l, "0123456789") {
		// For this to be a hostname, it must match a period
		// otherwise we found a port
		if l.Peek() != '.' {
			// This must be a port number, then
			l.Emit(ItemPortNumber)
			return next
		}
	}

	// Otherwise, slurp everything up to a ":"
	for {
		n := l.Next()
		switch n {
		case lex.EOF:
			l.EmitErrorf("Unexpected EOF")
			return nil
		case ':':
			l.Backup()
			l.Emit(ItemHostname)
			l.Next()
			l.Emit(ItemColon)
			if !lex.AcceptRun(l, "0123456789") {
				l.EmitErrorf("Expected number")
				return nil
			}
			l.Emit(ItemPortNumber)
			return next
		}
	}
}

type proxyRunner interface {
	Run(<-chan struct{}) error
}

func parseRule(r string) (proxyRunner, error) {
	p := &parser{}
	return p.Parse(r)
}

type parser struct {}

const (
	pSourceSection = iota
)

func (p *parser) Parse(s string) (proxyRunner, error) {
	l := lex.NewStringLexer(s, lexStart)
	go l.Run(l)

	ctx := &parseCtx{}

	if err := p.parseSource(l, ctx); err != nil {
		return nil, err
	}

	if err := p.parseDestination(l, ctx); err != nil {
		return nil, err
	}

	return p.composeRule(ctx), nil
}

func (p *parser) parseSource(l lex.Lexer, ctx *parseCtx) error {
	t := <-l.Items()

	switch t.Type() {
	case ItemHostname:
		src := &simpleHost{host: t.Value()}
		t = <-l.Items()
		if t.Type() != ItemPortNumber {
			return fmt.Errorf("Expected port")
		}
		src.port = t.Value()
		ctx.src = src
	case ItemPortNumber:
		src := &simpleHost{
			host: "0.0.0.0",
			port: t.Value(),
		}
		ctx.src = src
	default:
		return fmt.Errorf("Expected source <host:port>")
	}
	return nil
}

type parseCtx struct {
	src sourceSpec
	dst destinationSpec
}

type simpleHost struct {
	host string
	port string
}
type sourceSpec interface {}
type destinationSpec interface {}

func (p *parser) parseDestination(l lex.Lexer, ctx *parseCtx) error {
	// consume the arrow ->
	for {
		t := <-l.Items()
		if t.Type() == ItemWhitespace {
			continue
		}

		if t.Type() != ItemArrow {
			return fmt.Errorf("Expected '->', got %s", t.Type())
		}

		break
	}

	// Either a simple host or a block
	for ctx.dst == nil {
		t := <-l.Items()
		if t == nil {
			panic("Unexpeted end of input")
		}

		switch t.Type() {
		case ItemWhitespace:
			continue
		case ItemHostname:
			dst := &simpleHost{host: t.Value()}
			t = <-l.Items()
			if t.Type() != ItemPortNumber {
				return fmt.Errorf("Expected port")
			}
			dst.port = t.Value()
			ctx.dst = dst
		case ItemPortNumber:
			ctx.dst = &simpleHost{
				host: "0.0.0.0",
				port: t.Value(),
			}
		// case ItemRoundRobinBlock:
		// case ItemHashBlock:
		default:
			panic(fmt.Sprintf("Unknown type: %s", t.Type()))
		}
	}
	return nil
}

func (p *parser) composeRule(ctx *parseCtx) proxyRunner {
	switch t := ctx.dst.(type) {
	case *simpleHost:
		// XXX Assuming src = always simpleHost
		src := ctx.src.(*simpleHost)
		dst := ctx.dst.(*simpleHost)
		return NewProxy(
			fmt.Sprintf("%s:%s", src.host, src.port),
			fmt.Sprintf("%s:%s", dst.host, dst.port),
		)
	default:
		panic(fmt.Sprintf("Unknown type %s", t))
	}

	return nil
}
