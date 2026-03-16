package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

type Ratio struct {
	Key   string
	Value float64
}

// vsda implements the Vendor Slot Distribution Algorithm.
//
// How it works:
//
//  1. Remove vendors that have 0 db count and recalculate the ratio
//     based on the remaining vendors' total share.
//
//  2. Calculate the initial allocation for each vendor using the
//     available ratio and the given limit. Use ceiling to ensure
//     the total can reach the limit.
//
//  3. Sum the initial allocations and determine extras
//     (extras = totalAllocated - limit).
//
//  4. If extras exist, subtract from the highest priority vendor
//     first, then continue to lower priority vendors until the
//     extras are removed.
//
//  5. Check for exhausted vendors (where db count < allocated count)
//     and calculate the remaining slots.
//
//  6. Distribute the remaining slots to non-exhausted vendors based
//     on priority until all slots are filled or no vendors remain.
//
//  7. Return the final allocation as a list of vendor keys, ordered by priority.
//
// Constraints:
//
//	minimum allocation is 1, maximum is db count
//	vendors <= limit
func vsda(ratio map[string]float64, priority []string, dbcount map[string]int, limit int) []string {
	// These maps will hold the active vendors after filtering out those with 0 db count, and their corresponding ratios and db counts.
	activeRatio := make(map[string]float64)
	activeDB := make(map[string]int)
	removed := make(map[string]bool)
	var totalRatio float64

	// Remove that shares that has 0 db count, calculate based on remaining sum shares
	removedPriority := make(map[string]bool)
	for _, key := range priority {
		if dbcount[key] == 0 {
			removed[key] = true
			continue
		}
		activeRatio[key] = ratio[key]
		activeDB[key] = dbcount[key]
		totalRatio += ratio[key]
	}

	if len(activeRatio) > limit {
		fmt.Println("Error: Total vendors exceed the limit.")
		return nil
	}

	// remove vendor from priority if it is removed in previous step
	var updatedPriority []string
	for _, key := range priority {
		if !removedPriority[key] {
			updatedPriority = append(updatedPriority, key)
		}
	}

	// Calculate the initial counts from ratio and limit, and sum the total allocated count for all vendors.
	vendorCountTotal := 0
	vendorCount := make(map[string]int)
	for key, value := range activeRatio {
		c := int(math.Ceil((value / totalRatio) * float64(limit)))
		vendorCount[key] = c
		vendorCountTotal += c
	}

	// calculate extras after initial rounding
	extras := vendorCountTotal - limit

	// subtract from highest priority if extras are there and highest priority has more than 1 share
	if extras > 0 {
		for _, key := range updatedPriority {
			if vendorCount[key] > 1 {
				vendorCount[key]--
				extras--
				break
			}
		}
	}

	// untill extras are removed, subtract from lowest priority if it has more than 1 share, then move to higher priority
	for extras > 0 {
		for i := len(updatedPriority) - 1; i >= 0 && extras > 0; i-- {
			if vendorCount[updatedPriority[i]] > 1 {
				vendorCount[updatedPriority[i]]--
				extras--
			}
		}
	}

	// Calculate total spots remains and total allocations
	remainingSlots := 0
	exhaustedVendors := make(map[string]bool)
	for key, count := range vendorCount {
		if activeDB[key] <= count {
			remainingSlots += count - activeDB[key]
			vendorCount[key] = activeDB[key]
			exhaustedVendors[key] = true
		}
	}

	var queue []string

	for i := len(updatedPriority) - 1; i >= 0; i-- {
		key := updatedPriority[i]
		if !exhaustedVendors[key] {
			queue = append(queue, key)
		}
	}

	for remainingSlots > 0 && len(queue) > 0 {
		key := queue[0]
		queue = queue[1:]

		vendorCount[key]++
		remainingSlots--

		if vendorCount[key] < activeDB[key] {
			queue = append(queue, key)
		}
	}
	return result(vendorCount, updatedPriority)
}

func result(vendorCount map[string]int, priority []string) []string {
	var result []string

	for len(vendorCount) > 0 {
		for _, key := range priority {
			_, exists := vendorCount[key]
			if exists && vendorCount[key] > 0 {
				result = append(result, key)
				vendorCount[key]--
			}

			if vendorCount[key] == 0 {
				delete(vendorCount, key)
			}
		}
	}
	return result
}

func main() {
	testCases, err := loadTestCases("test_cases.json")
	if err != nil {
		fmt.Println("Error loading test cases:", err)
		return
	}

	for _, testCase := range testCases {
		fmt.Printf("Test Case: %s\n", testCase.Description)
		result := vsda(testCase.Ratio, testCase.Priority, testCase.DBCount, testCase.Limit)
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
