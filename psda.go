package main

import (
	"fmt"
	"math"
)

type Ratio struct {
	Key   string
	Value float64
}

// vsda implements the Property Slot Distribution Algorithm.
//
// How it works:
//
//  1. Remove partners that have 0 db count and recalculate the ratio
//     based on the remaining partners' total share.
//
//  2. Calculate the initial allocation for each vendor using the
//     available ratio and the given limit. Use ceiling to ensure
//     the total can reach the limit.
//
//  3. Sum the initial allocations and determine extras
//     (extras = totalAllocated - limit).
//
//  4. If extras exist, subtract from the highest priority vendor
//     first, then continue to lower priority partners until the
//     extras are removed.
//
//  5. Check for exhausted partners (where db count < allocated count)
//     and calculate the remaining slots.
//
//  6. Distribute the remaining slots to non-exhausted partners based
//     on priority until all slots are filled or no partners remain.
//
//  7. Return the final allocation as a list of vendor keys, ordered by priority.
//
// Constraints:
//
//	minimum allocation is 1, maximum is db count
//	partners <= limit
func psda(ratio map[string]float64, priority []string, dbcount map[string]int, limit int) []string {
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
		fmt.Println("Error: Total partners exceed the limit.")
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
	initialAllocationCount := 0
	partnerWiseCount := make(map[string]int)
	for key, value := range activeRatio {
		c := int(math.Ceil((value / totalRatio) * float64(limit)))
		partnerWiseCount[key] = c
		initialAllocationCount += c
	}

	// calculate extras after initial rounding
	extras := initialAllocationCount - limit

	// subtract from highest priority if extras are there and highest priority has more than 1 share
	if extras > 0 {
		for _, key := range updatedPriority {
			if partnerWiseCount[key] > 1 {
				partnerWiseCount[key]--
				extras--
				break
			}
		}
	}

	// untill extras are removed, subtract from lowest priority if it has more than 1 share, then move to higher priority
	for extras > 0 {
		for i := len(updatedPriority) - 1; i >= 0 && extras > 0; i-- {
			if partnerWiseCount[updatedPriority[i]] > 1 {
				partnerWiseCount[updatedPriority[i]]--
				extras--
			}
		}
	}

	// Calculate total spots remains and total allocations
	remainingSlots := 0
	exhaustedVendors := make(map[string]bool)
	for key, count := range partnerWiseCount {
		if activeDB[key] <= count {
			remainingSlots += count - activeDB[key]
			partnerWiseCount[key] = activeDB[key]
			exhaustedVendors[key] = true
		}
	}

	var queue []string

	if remainingSlots > 0 {
		for i := len(updatedPriority) - 1; i >= 0; i-- {
			key := updatedPriority[i]
			if !exhaustedVendors[key] {
				queue = append(queue, key)
			}
		}

		for remainingSlots > 0 && len(queue) > 0 {
			key := queue[0]
			queue = queue[1:]

			partnerWiseCount[key]++
			remainingSlots--

			if partnerWiseCount[key] < activeDB[key] {
				queue = append(queue, key)
			}
		}
	}

	return generateSequence(partnerWiseCount, updatedPriority)
}

func generateSequence(partnerWiseCount map[string]int, priority []string) []string {
	var result []string

	for len(partnerWiseCount) > 0 {
		for _, key := range priority {
			_, exists := partnerWiseCount[key]
			if exists && partnerWiseCount[key] > 0 {
				result = append(result, key)
				partnerWiseCount[key]--
			}

			if partnerWiseCount[key] == 0 {
				delete(partnerWiseCount, key)
			}
		}
	}
	return result
}
