// Package calculator evaluates arithmetic expressions without delegating to a
// shell or general-purpose scripting runtime.
package calculator

import (
	"fmt"
	"math"
	"strconv"
	"unicode"
)

// Evaluate parses and evaluates an arithmetic expression. Supported operators
// are +, -, *, /, %, ^, parentheses, and unary plus/minus.
func Evaluate(expression string) (float64, error) {
	p := parser{input: []rune(expression)}
	value, err := p.parseExpression()
	if err != nil {
		return 0, err
	}
	p.skipSpace()
	if !p.atEnd() {
		return 0, p.errorf("unexpected %q", p.input[p.position])
	}
	return value, nil
}

type parser struct {
	input    []rune
	position int
}

func (p *parser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpace()
		if !p.consume('+') && !p.consume('-') {
			return left, nil
		}
		operator := p.input[p.position-1]
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if operator == '+' {
			left += right
		} else {
			left -= right
		}
	}
}

func (p *parser) parseTerm() (float64, error) {
	left, err := p.parsePower()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpace()
		if !p.consume('*') && !p.consume('/') && !p.consume('%') {
			return left, nil
		}
		operator := p.input[p.position-1]
		right, err := p.parsePower()
		if err != nil {
			return 0, err
		}
		switch operator {
		case '*':
			left *= right
		case '/':
			if right == 0 {
				return 0, p.errorf("division by zero")
			}
			left /= right
		case '%':
			if right == 0 {
				return 0, p.errorf("remainder by zero")
			}
			left = math.Mod(left, right)
		}
	}
}

func (p *parser) parsePower() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	p.skipSpace()
	if !p.consume('^') {
		return left, nil
	}
	right, err := p.parsePower()
	if err != nil {
		return 0, err
	}
	return math.Pow(left, right), nil
}

func (p *parser) parseUnary() (float64, error) {
	p.skipSpace()
	if p.consume('+') {
		return p.parseUnary()
	}
	if p.consume('-') {
		value, err := p.parseUnary()
		return -value, err
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (float64, error) {
	p.skipSpace()
	if p.consume('(') {
		value, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		p.skipSpace()
		if !p.consume(')') {
			return 0, p.errorf("expected closing parenthesis")
		}
		return value, nil
	}
	return p.parseNumber()
}

func (p *parser) parseNumber() (float64, error) {
	p.skipSpace()
	start := p.position
	digits := false
	for !p.atEnd() && unicode.IsDigit(p.input[p.position]) {
		digits = true
		p.position++
	}
	if !p.atEnd() && p.input[p.position] == '.' {
		p.position++
		for !p.atEnd() && unicode.IsDigit(p.input[p.position]) {
			digits = true
			p.position++
		}
	}
	if !digits {
		if p.atEnd() {
			return 0, p.errorf("expected a number")
		}
		return 0, p.errorf("expected a number, found %q", p.input[p.position])
	}
	value, err := strconv.ParseFloat(string(p.input[start:p.position]), 64)
	if err != nil {
		return 0, p.errorf("invalid number")
	}
	return value, nil
}

func (p *parser) skipSpace() {
	for !p.atEnd() && unicode.IsSpace(p.input[p.position]) {
		p.position++
	}
}

func (p *parser) consume(expected rune) bool {
	if p.atEnd() || p.input[p.position] != expected {
		return false
	}
	p.position++
	return true
}

func (p *parser) atEnd() bool { return p.position >= len(p.input) }

func (p *parser) errorf(format string, args ...any) error {
	return fmt.Errorf("at character %d: %s", p.position+1, fmt.Sprintf(format, args...))
}
