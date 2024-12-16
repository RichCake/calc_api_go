package calculation

import (
	"strconv"
	"strings"
	"unicode"
)

func Calc(expression string) (float64, error) {
	rpn, err := toPostfix(expression)
	if err != nil {
		return 0, err
	}
	return evalPostfix(rpn)
}

func toPostfix(expression string) ([]string, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	var output []string
	var stack []string
	priority := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	for i := 0; i < len(expression); i++ {
		char := string(expression[i])
		if ("0" <= char && char <= "9") || char == "." || 
			(char == "-" && (i == 0 || string(expression[i-1]) == "(" || priority[string(expression[i-1])] > 0)) {
			number := char
			for i+1 < len(expression) && (unicode.IsDigit(rune(expression[i+1])) || string(expression[i+1]) == ".") {
				i++
				number += string(expression[i])
			}
			output = append(output, number)
		} else if char == "(" {
			stack = append(stack, char)
		} else if char == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 || stack[len(stack)-1] != "(" {
				return nil, ErrMismatchedBracket
			}
			stack = stack[:len(stack)-1]
		} else if priority[char] > 0 {
			for len(stack) > 0 && priority[stack[len(stack)-1]] >= priority[char] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, char)
		} else {
			return nil, ErrInvalidSymbols
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, ErrMismatchedBracket
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

func evalPostfix(rpn []string) (float64, error) {
	var stack []float64
	for _, value := range rpn {
		if n, err := strconv.ParseFloat(value, 64); err == nil {
			stack = append(stack, n)
		} else {
			if len(stack) < 2 {
				return 0, ErrInvalidOperationsPlacement
			}
			n2, n1 := stack[len(stack)-1], stack[len(stack)-2]
			var res float64
			switch value {
			case "+":
				res = n1 + n2
			case "-":
				res = n1 - n2
			case "*":
				res = n1 * n2
			case "/":
				if n2 == 0 {
					return 0, ErrZeroDivision
				}
				res = n1 / n2
			}
			stack = stack[:len(stack)-2]
			stack = append(stack, res)
		}
	}
	if len(stack) != 1 {
		return 0, ErrInvalidExpression
	}
	return stack[0], nil
}
