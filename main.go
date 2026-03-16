package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	testCases, err := loadTestCases("test_cases.json")
	if err != nil {
		fmt.Println("Error loading test cases:", err)
		return
	}

	for _, testCase := range testCases {
		fmt.Printf("Test Case: %s\n", testCase.Description)
		result := psda(testCase.Ratio, testCase.Priority, testCase.DBCount, testCase.Limit)
		fmt.Println("Limit:", testCase.Limit)
		fmt.Printf("Ratio: %v\n", testCase.Ratio)
		fmt.Printf("Priority: %v\n", testCase.Priority)
		fmt.Printf("DB Count: %v\n", testCase.DBCount)
		fmt.Printf("Result: %v\n", result)
		fmt.Println("--------------------------------------------------")
	}
}

type TestCase struct {
	Description string
	Ratio       map[string]float64
	Priority    []string
	DBCount     map[string]int
	Limit       int
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
