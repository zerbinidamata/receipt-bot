package recipe

import (
	"testing"
	"time"
)

func TestNewInstruction(t *testing.T) {
	duration := 5 * time.Minute

	tests := []struct {
		name        string
		stepNumber  int
		text        string
		duration    *time.Duration
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid instruction",
			stepNumber: 1,
			text:       "Preheat oven to 350°F",
			duration:   nil,
			wantErr:    false,
		},
		{
			name:       "with duration",
			stepNumber: 2,
			text:       "Bake for 5 minutes",
			duration:   &duration,
			wantErr:    false,
		},
		{
			name:        "invalid step number zero",
			stepNumber:  0,
			text:        "Mix ingredients",
			duration:    nil,
			wantErr:     true,
			errContains: "step number must be positive",
		},
		{
			name:        "invalid step number negative",
			stepNumber:  -1,
			text:        "Mix ingredients",
			duration:    nil,
			wantErr:     true,
			errContains: "step number must be positive",
		},
		{
			name:        "empty text",
			stepNumber:  1,
			text:        "",
			duration:    nil,
			wantErr:     true,
			errContains: "text cannot be empty",
		},
		{
			name:       "whitespace text trimmed",
			stepNumber: 1,
			text:       "  Mix ingredients  ",
			duration:   nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst, err := NewInstruction(tt.stepNumber, tt.text, tt.duration)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewInstruction() expected error but got nil")
					return
				}
				if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("NewInstruction() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewInstruction() unexpected error = %v", err)
				return
			}

			if inst.StepNumber() != tt.stepNumber {
				t.Errorf("StepNumber() = %v, want %v", inst.StepNumber(), tt.stepNumber)
			}

			if inst.Text() == "" {
				t.Errorf("Text() is empty after trimming")
			}
		})
	}
}

func TestInstruction_String(t *testing.T) {
	fiveMin := 5 * time.Minute

	tests := []struct {
		name     string
		inst     Instruction
		expected string
	}{
		{
			name: "without duration",
			inst: Instruction{
				stepNumber: 1,
				text:       "Preheat oven",
				duration:   nil,
			},
			expected: "1. Preheat oven",
		},
		{
			name: "with duration",
			inst: Instruction{
				stepNumber: 2,
				text:       "Bake",
				duration:   &fiveMin,
			},
			expected: "2. Bake (5 min)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.inst.String()
			if got != tt.expected {
				t.Errorf("Instruction.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
