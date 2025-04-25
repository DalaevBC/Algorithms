package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func tokenize(input string) []string {
	re := regexp.MustCompile(`\(|\)|\S+`)
	return re.FindAllString(input, -1)
}

func parse(tokens []string) (interface{}, []string) {
	if len(tokens) == 0 {
		return nil, tokens
	}

	token := tokens[0]
	tokens = tokens[1:]

	if token == "(" {
		var list []interface{}
		for len(tokens) > 0 && tokens[0] != ")" {
			var elem interface{}
			elem, tokens = parse(tokens)
			list = append(list, elem)
		}
		if len(tokens) > 0 && tokens[0] == ")" {
			tokens = tokens[1:]
		}
		return list, tokens
	} else {
		return token, tokens
	}
}

func isConstant(expr interface{}) bool {
	switch e := expr.(type) {
	case string:
		return e != "x"
	case []interface{}:
		for _, arg := range e[1:] {
			if !isConstant(arg) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func differentiate(expr interface{}) interface{} {
	switch e := expr.(type) {
	case string:
		if e == "x" {
			return "1"
		} else {
			if _, err := strconv.ParseFloat(e, 64); err == nil {
				return "0"
			}
			return "0"
		}
	case []interface{}:
		if len(e) == 0 {
			return "0"
		}

		op, _ := e[0].(string)
		args := e[1:]

		switch op {
		case "+":
			da := differentiate(args[0])
			db := differentiate(args[1])
			return []interface{}{"+", da, db}
		case "-":
			da := differentiate(args[0])
			db := differentiate(args[1])
			return []interface{}{"-", da, db}
		case "*":
			a := args[0]
			b := args[1]
			da := differentiate(a)
			db := differentiate(b)
			return []interface{}{"+", []interface{}{"*", da, b}, []interface{}{"*", a, db}}
		case "/":
			a := args[0]
			b := args[1]
			da := differentiate(a)
			db := differentiate(b)
			numerator := []interface{}{"-", []interface{}{"*", da, b}, []interface{}{"*", a, db}}
			denominator := []interface{}{"^", b, "2"}
			return []interface{}{"/", numerator, denominator}
		case "^":
			a := args[0]
			b := args[1]
			da := differentiate(a)
			db := differentiate(b)
			if isConstant(b) {
				exponentMinus1 := []interface{}{"-", b, "1"}
				part1 := []interface{}{"*", b, []interface{}{"^", a, exponentMinus1}}
				return []interface{}{"*", part1, da}
			} else {
				term1 := []interface{}{"/", []interface{}{"*", b, da}, a}
				term2 := []interface{}{"*", db, []interface{}{"ln", a}}
				sumTerms := []interface{}{"+", term1, term2}
				return []interface{}{"*", []interface{}{"^", a, b}, sumTerms}
			}
		case "cos":
			u := args[0]
			du := differentiate(u)
			return []interface{}{"*", []interface{}{"*", "-1", []interface{}{"sin", u}}, du}
		case "sin":
			u := args[0]
			du := differentiate(u)
			return []interface{}{"*", []interface{}{"cos", u}, du}
		case "tan":
			u := args[0]
			du := differentiate(u)
			tanU := []interface{}{"tan", u}
			tanSq := []interface{}{"^", tanU, "2"}
			onePlus := []interface{}{"+", "1", tanSq}
			return []interface{}{"*", onePlus, du}
		case "exp":
			u := args[0]
			du := differentiate(u)
			return []interface{}{"*", []interface{}{"exp", u}, du}
		case "ln":
			u := args[0]
			du := differentiate(u)
			return []interface{}{"*", []interface{}{"/", "1", u}, du}
		default:
			panic("unknown operator: " + op)
		}
	default:
		return "0"
	}
}

func simplify(expr interface{}) interface{} {
	switch e := expr.(type) {
	case string:
		return e
	case []interface{}:
		if len(e) == 0 {
			return e
		}

		simplified := make([]interface{}, len(e))
		for i, arg := range e {
			simplified[i] = simplify(arg)
		}

		op, ok := simplified[0].(string)
		if !ok {
			return simplified
		}

		switch op {
		case "+", "-", "*", "/", "^":
			if len(simplified) != 3 {
				return simplified
			}
			a := simplified[1]
			b := simplified[2]

			aNum, aIsNum := isNumber(a)
			bNum, bIsNum := isNumber(b)

			if aIsNum && bIsNum {
				switch op {
				case "+":
					return fmt.Sprintf("%g", aNum+bNum)
				case "-":
					return fmt.Sprintf("%g", aNum-bNum)
				case "*":
					return fmt.Sprintf("%g", aNum*bNum)
				case "/":
					if bNum != 0 {
						return fmt.Sprintf("%g", aNum/bNum)
					}
				case "^":
					return fmt.Sprintf("%g", math.Pow(aNum, bNum))
				}
			}

			switch op {
			case "+":
				if isZero(a) {
					return b
				}
				if isZero(b) {
					return a
				}
			case "-":
				if isZero(b) {
					return a
				}
			case "*":
				if isZero(a) || isZero(b) {
					return "0"
				}
				if isOne(a) {
					return b
				}
				if isOne(b) {
					return a
				}
			case "/":
				if isZero(a) {
					return "0"
				}
				if isOne(b) {
					return a
				}
			case "^":
				if isZero(b) {
					return "1"
				}
				if isOne(b) {
					return a
				}
			}

			return []interface{}{op, a, b}
		default:
			return simplified
		}
	default:
		return expr
	}
}

func isNumber(s interface{}) (float64, bool) {
	str, ok := s.(string)
	if !ok {
		return 0, false
	}
	f, err := strconv.ParseFloat(str, 64)
	return f, err == nil
}

func isZero(s interface{}) bool {
	if str, ok := s.(string); ok {
		return str == "0" || str == "0.0"
	}
	f, ok := isNumber(s)
	return ok && f == 0.0
}

func isOne(s interface{}) bool {
	if str, ok := s.(string); ok {
		return str == "1" || str == "1.0"
	}
	f, ok := isNumber(s)
	return ok && f == 1.0
}

func exprToString(expr interface{}) string {
	switch e := expr.(type) {
	case string:
		return e
	case []interface{}:
		parts := make([]string, len(e))
		for i, elem := range e {
			parts[i] = exprToString(elem)
		}
		return "(" + strings.Join(parts, " ") + ")"
	default:
		return ""
	}
}

func main() {
	var input string
	fmt.Scanln(&input)
	tokens := tokenize(input)
	expr, _ := parse(tokens)
	derivative := differentiate(expr)
	simplified := simplify(derivative)
	output := exprToString(simplified)
	fmt.Println(output)
}