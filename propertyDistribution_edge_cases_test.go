package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"
)

type TestCase struct {
	Case        int                `json:"case"`
	Description string             `json:"description"`
	Ratio       map[string]float64 `json:"ratio"`
	Priority    []string           `json:"priority"`
	DBCount     map[string]int     `json:"dbCount"`
	Limit       int                `json:"limit"`
	Expected    []string           `json:"expected"`
}

func loadTestCases(path string) ([]TestCase, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tests []TestCase
	if err := json.Unmarshal(file, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}

func TestPropertyDistributionEdgeCases(t *testing.T) {
	testCases, err := loadTestCases("test_cases.json")
	if err != nil {
		fmt.Println("Error loading test cases:", err)
		return
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %d: %s", testCase.Case, testCase.Description), func(t *testing.T) {
			got := propertyDistribution(testCase.Ratio, testCase.Priority, testCase.DBCount, testCase.Limit)
			want := testCase.Expected
			if !slices.Equal(got, want) {
				t.Logf("got %v, want %v", got, want)
			}
		})
	}
}
