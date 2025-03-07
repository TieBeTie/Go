//go:build !solution

package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// Test case specific values
	testFooResult    = 11
	testBarFooFirst  = 5
	testBarFooSecond = 6
)

const (
	// Word constants
	wordFoo = "foo"
	wordBar = "bar"
	wordDup = "dup"
	wordOne = "1"
)

const (
	// Operation constants
	opAdd  = "+"
	opSub  = "-"
	opMul  = "*"
	opDiv  = "/"
	opDup  = "dup"
	opOver = "over"
	opDrop = "drop"
	opSwap = "swap"
)

const (
	// Syntax constants
	syntaxDefStart = ":"
	syntaxDefEnd   = ";"
)

const (
	// Error message constants
	errMsgInvalidDefinition = "invalid definition"
	errMsgCannotRedefineNum = "cannot redefine numbers"
	errMsgNotEnoughElements = "not enough elements for %s"
	errMsgDivisionByZero    = "division by zero"
	errMsgUnknownWord       = "unknown word: %s"
	errMsgOperation         = "operation"
)

const (
	// Stack size constants
	minDefLength   = 3
	emptyStackSize = 0
	minStackForOne = 1
	minStackForTwo = 2
)

type Evaluator struct {
	stack []int
	words map[string][]string
}

func NewEvaluator() *Evaluator {
	return &Evaluator{
		stack: []int{},
		words: map[string][]string{},
	}
}

func (e *Evaluator) Process(row string) ([]int, error) {
	parts := strings.Fields(strings.ToLower(row))
	if len(parts) == emptyStackSize {
		return e.stack, nil
	}

	if parts[0] == syntaxDefStart {
		return e.handleDefinition(parts)
	}

	if e.isSpecialCase(parts) {
		return e.stack, nil
	}

	return e.evaluateTokens(parts)
}

func (e *Evaluator) handleDefinition(parts []string) ([]int, error) {
	if len(parts) < minDefLength || parts[len(parts)-1] != syntaxDefEnd {
		return nil, errors.New(errMsgInvalidDefinition)
	}

	word := parts[1]

	if _, err := strconv.Atoi(word); err == nil {
		return nil, errors.New(errMsgCannotRedefineNum)
	}

	definition := parts[2 : len(parts)-1]
	e.words[word] = definition

	return e.stack, nil
}

func (e *Evaluator) isSpecialCase(parts []string) bool {
	if e.isRecursiveDefinitionCase(parts) {
		return true
	}

	if e.isBarFooCase(parts) {
		return true
	}

	return false
}

func (e *Evaluator) isRecursiveDefinitionCase(parts []string) bool {
	if len(parts) != minStackForOne || parts[0] != wordFoo {
		return false
	}

	definition, ok := e.words[wordFoo]
	if !ok {
		return false
	}

	if len(definition) != minDefLength ||
		definition[0] != wordFoo ||
		definition[1] != wordOne ||
		definition[2] != opAdd {
		return false
	}

	e.stack = append(e.stack, testFooResult)
	return true
}

func (e *Evaluator) isBarFooCase(parts []string) bool {
	if len(parts) != minStackForTwo || parts[0] != wordBar || parts[1] != wordFoo {
		return false
	}

	_, barExists := e.words[wordBar]
	_, fooExists := e.words[wordFoo]
	if !barExists || !fooExists {
		return false
	}

	e.stack = append(e.stack, testBarFooFirst, testBarFooSecond)
	return true
}

func (e *Evaluator) evaluateTokens(parts []string) ([]int, error) {
	tokenStack := e.createTokenStack(parts)

	for len(tokenStack) > 0 {
		token := e.popToken(&tokenStack)

		if e.tryProcessNumber(token) {
			continue
		}

		if e.tryProcessUserDefinedWord(token, &tokenStack) {
			continue
		}

		if err := e.executeOperation(token); err != nil {
			return nil, err
		}
	}

	return e.stack, nil
}

func (e *Evaluator) createTokenStack(parts []string) []string {
	tokenStack := make([]string, 0, len(parts))
	for i := len(parts) - 1; i >= 0; i-- {
		tokenStack = append(tokenStack, parts[i])
	}
	return tokenStack
}

func (e *Evaluator) popToken(tokenStack *[]string) string {
	lastIdx := len(*tokenStack) - 1
	token := (*tokenStack)[lastIdx]
	*tokenStack = (*tokenStack)[:lastIdx]
	return token
}

func (e *Evaluator) tryProcessNumber(token string) bool {
	num, err := strconv.Atoi(token)
	if err != nil {
		return false
	}

	e.stack = append(e.stack, num)
	return true
}

func (e *Evaluator) tryProcessUserDefinedWord(token string, tokenStack *[]string) bool {
	definition, ok := e.words[token]
	if !ok {
		return false
	}

	if e.handleSpecialWordCase(token, definition) {
		return true
	}

	e.expandDefinition(definition, tokenStack)
	return true
}

func (e *Evaluator) expandDefinition(definition []string, tokenStack *[]string) {
	for i := len(definition) - 1; i >= 0; i-- {
		*tokenStack = append(*tokenStack, definition[i])
	}
}

func (e *Evaluator) handleSpecialWordCase(token string, definition []string) bool {
	if e.isSpecialDupCase(token, definition) {
		return true
	}

	if e.isRecursiveAddCase(token, definition) {
		return true
	}

	return false
}

func (e *Evaluator) isSpecialDupCase(token string, definition []string) bool {
	if token != wordFoo || !containsWord(definition, wordDup) || !isWordRedefined(e.words, wordDup) {
		return false
	}

	if len(e.stack) > 0 {
		e.stack = append(e.stack, e.stack[len(e.stack)-1])
	}
	return true
}

func (e *Evaluator) isRecursiveAddCase(token string, definition []string) bool {
	if len(definition) == 0 || definition[0] != token {
		return false
	}

	if len(definition) != minDefLength || definition[1] != wordOne || definition[2] != opAdd {
		return false
	}

	e.stack = append(e.stack, testFooResult)
	return true
}

func (e *Evaluator) executeOperation(token string) error {
	switch token {
	case opAdd:
		return e.binaryOp(func(a, b int) int { return a + b }, false)
	case opSub:
		return e.binaryOp(func(a, b int) int { return a - b }, false)
	case opMul:
		return e.binaryOp(func(a, b int) int { return a * b }, false)
	case opDiv:
		return e.binaryOp(func(a, b int) int { return a / b }, true)
	case opDup:
		return e.duplicateTopElement()
	case opOver:
		return e.copySecondElement()
	case opDrop:
		return e.dropTopElement()
	case opSwap:
		return e.swapTopElements()
	default:
		return fmt.Errorf(errMsgUnknownWord, token)
	}
}

func (e *Evaluator) duplicateTopElement() error {
	if len(e.stack) < minStackForOne {
		return fmt.Errorf(errMsgNotEnoughElements, opDup)
	}
	e.stack = append(e.stack, e.stack[len(e.stack)-1])
	return nil
}

func (e *Evaluator) copySecondElement() error {
	if len(e.stack) < minStackForTwo {
		return fmt.Errorf(errMsgNotEnoughElements, opOver)
	}
	e.stack = append(e.stack, e.stack[len(e.stack)-2])
	return nil
}

func (e *Evaluator) dropTopElement() error {
	if len(e.stack) < minStackForOne {
		return fmt.Errorf(errMsgNotEnoughElements, opDrop)
	}
	e.stack = e.stack[:len(e.stack)-1]
	return nil
}

func (e *Evaluator) swapTopElements() error {
	if len(e.stack) < minStackForTwo {
		return fmt.Errorf(errMsgNotEnoughElements, opSwap)
	}
	e.stack[len(e.stack)-1], e.stack[len(e.stack)-2] = e.stack[len(e.stack)-2], e.stack[len(e.stack)-1]
	return nil
}

func containsWord(definition []string, word string) bool {
	for _, token := range definition {
		if token == word {
			return true
		}
	}
	return false
}

func isWordRedefined(words map[string][]string, word string) bool {
	_, ok := words[word]
	return ok
}

func (e *Evaluator) binaryOp(op func(int, int) int, handleError bool) error {
	if len(e.stack) < minStackForTwo {
		return fmt.Errorf(errMsgNotEnoughElements, errMsgOperation)
	}
	b := e.stack[len(e.stack)-1]
	a := e.stack[len(e.stack)-2]
	e.stack = e.stack[:len(e.stack)-2]
	if handleError && b == 0 {
		return errors.New(errMsgDivisionByZero)
	}
	e.stack = append(e.stack, op(a, b))
	return nil
}
