package recipe

import (
	"fmt"
	"receipt-bot/internal/domain/shared"
	"strings"
	"time"
)

// Instruction represents a cooking instruction step (Value Object)
type Instruction struct {
	stepNumber int
	text       string
	duration   *time.Duration
}

// NewInstruction creates a new Instruction
func NewInstruction(stepNumber int, text string, duration *time.Duration) (Instruction, error) {
	text = strings.TrimSpace(text)

	if stepNumber <= 0 {
		return Instruction{}, shared.ErrInvalidStepNumber
	}

	if text == "" {
		return Instruction{}, shared.ErrInvalidInstructionText
	}

	return Instruction{
		stepNumber: stepNumber,
		text:       text,
		duration:   duration,
	}, nil
}

// StepNumber returns the instruction step number
func (i Instruction) StepNumber() int {
	return i.stepNumber
}

// Text returns the instruction text
func (i Instruction) Text() string {
	return i.text
}

// Duration returns the instruction duration
func (i Instruction) Duration() *time.Duration {
	return i.duration
}

// String returns a formatted string representation
func (i Instruction) String() string {
	result := fmt.Sprintf("%d. %s", i.stepNumber, i.text)
	if i.duration != nil {
		minutes := int(i.duration.Minutes())
		if minutes > 0 {
			result += fmt.Sprintf(" (%d min)", minutes)
		}
	}
	return result
}
