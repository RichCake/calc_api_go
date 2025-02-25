package calculation

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Tree struct {
	Root *TreeNode
}

func PrintTree(node *TreeNode, level int) {
	if node == nil {
		return
	}
	PrintTree(node.Right, level+1) // Сначала правый потомок
	fmt.Printf("%s%s\n", strings.Repeat("  ", level), node.Val)
	PrintTree(node.Left, level+1) // Потом левый потомок
}

type TreeNode struct {
	Val    string
	Left   *TreeNode
	Right  *TreeNode
	TaskID int
}

func (node *TreeNode) IsSpare() bool {
	if node.Right != nil && node.Left != nil {
		_, err1 := strconv.ParseFloat(node.Left.Val, 64)
		_, err2 := strconv.ParseFloat(node.Right.Val, 64)
		if err1 == nil && err2 == nil {
			return true
		}
	}
	return false
}

func (t *Tree) FindSpareNodes() []*TreeNode {
	spare_nodes := []*TreeNode{}
	stack := []*TreeNode{t.Root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node.IsSpare() {
			spare_nodes = append(spare_nodes, node)
		} else {
			if node.Right != nil {
				stack = append(stack, node.Right)
			}
			if node.Left != nil {
				stack = append(stack, node.Left)
			}
		}
	}
	return spare_nodes
}

func (t *Tree) ReplaceNodeWithValue(node *TreeNode, val float64) {
	node.Left = nil
	node.Right = nil
	arg := strconv.FormatFloat(val, 'f', -1, 64)
	node.Val = arg
}

func (t *Tree) FindParentAndNodeByTaskID(task_id int) (*TreeNode, *TreeNode) {
	if t.Root.TaskID == task_id {
		return nil, t.Root
	}

	stack := []*TreeNode{t.Root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node.Right != nil && node.Right.TaskID == task_id {
			return node, node.Right
		}
		if node.Left != nil && node.Left.TaskID == task_id {
			return node, node.Left
		}

		// Добавляем только непустые узлы
		if node.Right != nil {
			stack = append(stack, node.Right)
		}
		if node.Left != nil {
			stack = append(stack, node.Left)
		}

	}
	return nil, nil
}

func ToPostfix(expression string) ([]string, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	var output []string
	var stack []string

	priority := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	// Проверка на пустое выражение
	if len(expression) == 0 {
		return nil, ErrInvalidExpression
	}

	var prevToken string // Хранит предыдущий обработанный токен

	for i := 0; i < len(expression); i++ {
		char := string(expression[i])

		// Число (включая десятичные дроби)
		if unicode.IsDigit(rune(expression[i])) || char == "." ||
			(char == "-" && (i == 0 || prevToken == "(" || priority[prevToken] > 0)) {

			number := char
			for i+1 < len(expression) && (unicode.IsDigit(rune(expression[i+1])) || string(expression[i+1]) == ".") {
				i++
				number += string(expression[i])
			}
			output = append(output, number)
			prevToken = number

		} else if char == "(" {
			stack = append(stack, char)
			prevToken = char

		} else if char == ")" {
			// Перед `)` должно быть число или `)`
			if prevToken == "" || priority[prevToken] > 0 || prevToken == "(" {
				return nil, ErrMismatchedBracket
			}

			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}

			if len(stack) == 0 {
				return nil, ErrMismatchedBracket
			}

			stack = stack[:len(stack)-1] // Удаляем `(`
			prevToken = ")"

		} else if priority[char] > 0 {
			// Запрещаем два оператора подряд
			if prevToken == "" || priority[prevToken] > 0 || prevToken == "(" {
				return nil, ErrInvalidOperationsPlacement
			}

			for len(stack) > 0 && priority[stack[len(stack)-1]] >= priority[char] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, char)
			prevToken = char

		} else {
			return nil, ErrInvalidSymbols
		}
	}

	// Проверка на завершение выражения
	if prevToken == "" || priority[prevToken] > 0 {
		return nil, ErrInvalidExpression
	}

	// Выгружаем оставшиеся операторы из стека
	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, ErrMismatchedBracket
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

func BuildTree(postfix []string) *Tree {
	stack := []*TreeNode{} // Используем указатели

	for _, token := range postfix {
		_, err := strconv.ParseFloat(token, 64)
		if err == nil {
			// Если это число, создаем узел и кладем в стек
			stack = append(stack, &TreeNode{Val: token})
		} else {
			// Если это оператор, забираем два верхних операнда из стека
			if len(stack) < 2 {
				panic("Invalid expression: not enough operands")
			}
			right := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			left := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			// Создаем узел оператора с двумя потомками
			node := &TreeNode{Val: token, Left: left, Right: right}
			stack = append(stack, node)
		}
	}

	if len(stack) != 1 {
		panic("Invalid expression: leftover nodes in stack")
	}

	return &Tree{Root: stack[0]}
}
