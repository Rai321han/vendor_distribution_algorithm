package main

import (
	"math"
)

type Ratio struct {
	Key   string
	Value float64
}

// propertyDistribution is a function that distributes a limited number of slots among partners based on their specified ratios and priorities, while also considering the available slots for each partner in the database.
//
// It takes a ratio map (partner key to ratio), a priority list of partner keys, a db count map (partner key to available slots), and a limit on total slots to distribute.
//
// It returns a list of partner keys representing the allocated slots, ordered by priority.
//
// example usage:
//
//	ratio := map[string]float64{"partnerA": 50, "partnerB": 30, "partnerC": 20}
//	priority := []string{"partnerA", "partnerB", "partnerC"}
//	dbcount := map[string]int{"partnerA": 5, "partnerB": 3, "partnerC": 2}
//	limit := 7
//	result := propertyDistribution(ratio, priority, dbcount, limit)
//
// How it works:
//
//  1. Remove partners that have 0 db count and recalculate the ratio
//     based on the remaining partners' total share.
//
//  2. Calculate the initial allocation for each partner using the
//     available ratio and the given limit.
//
//  3. Sum the initial allocations and determine extras
//     (extras = totalAllocated - limit).
//
//  4. If extras exist, subtract from the highest priority partner first, if highest priority partner has 1, then subtract from next higher and so on. If extras still remains or all partner has 1 allocaiton, then continue subtraction from lowest priority to highest partners until the
//     extras are removed.
//
//  5. Check for exhausted partners (where db count < allocated count)
//     and calculate the remaining slots.
//
//  6. Distribute the remaining slots to non-exhausted partners based
//     on priority until all slots are filled or no partners remain.
//
//  7. If total allocated count exceeds the limit, remove in round robin manner from lowest priority to highest until total allocated count is equal to limit.
//
//  8. Return the final allocation as a list of partner keys, ordered by priority.
//
// Constraints:
//
//	minimum allocation is 1, maximum is db count
//	1 <= limit
//	dbcount[partner] >= 0
//	ratio[partner] >= 0
func propertyDistribution(ratio map[string]float64, priority []string, dbcount map[string]int, limit int) []string {

	// length of ratio, priority and dbcount should be same.
	ratioLength := len(ratio)
	priorityLength := len(priority)
	dbcountLength := len(dbcount)

	if ratioLength != priorityLength || ratioLength != dbcountLength {
		return []string{}
	}

	// These maps will hold the active partners after filtering out those with 0 db count, and their corresponding ratios and db counts.
	activeRatio := make(map[string]float64)
	activeDB := make(map[string]int)
	removed := make(map[string]bool)
	var totalRatio float64

	// Remove that shares that has 0 db count, calculate based on remaining sum shares
	for _, key := range priority {
		if dbcount[key] == 0 {
			removed[key] = true
			continue
		}
		activeRatio[key] = ratio[key]
		activeDB[key] = dbcount[key]
		totalRatio += ratio[key]
	}

	// remove partner from priority if it is removed in previous step
	var updatedPriority []string
	for _, key := range priority {
		if !removed[key] {
			updatedPriority = append(updatedPriority, key)
		}
	}

	// Calculate the initial counts from ratio and limit, and sum the total allocated count for all partners.
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
		// if highest priority partner has only 1 then move toward next higher
		for _, key := range updatedPriority {
			if partnerWiseCount[key] > 1 {
				partnerWiseCount[key]--
				extras--
				break
			}
		}
	}

	// untill extras are removed, subtract from lowest priority if it has more than 1 share, then move to higher priority
	exhaustedAfterRemoved := len(partnerWiseCount)
	for extras > 0 && exhaustedAfterRemoved > 0 {
		for i := len(updatedPriority) - 1; i >= 0 && extras > 0; i-- {
			if partnerWiseCount[updatedPriority[i]] > 1 {
				partnerWiseCount[updatedPriority[i]]--
				extras--
			}

			if partnerWiseCount[updatedPriority[i]] == 1 {
				exhaustedAfterRemoved--
			}
		}
	}

	// Calculate total spots remains and total allocations
	remainingSlots := 0
	exhaustedpartners := make(map[string]bool)
	for key, count := range partnerWiseCount {
		if activeDB[key] <= count {
			remainingSlots += count - activeDB[key]
			partnerWiseCount[key] = activeDB[key]
			exhaustedpartners[key] = true
		}
	}

	var queue []string

	// Distribute remaining slots to non-exhausted partners based on priority until all slots are filled or no partners remain.
	if remainingSlots > 0 {
		for i := len(updatedPriority) - 1; i >= 0; i-- {
			key := updatedPriority[i]
			if !exhaustedpartners[key] {
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

	// limit < partners
	// only possible when all partners have at most 1 allocaiton
	// in that case we will remove in round robin manner from lowest priority to highest until total allocated count is equal to limit.
	// that means we need to remove (partners - limit) partners from the parterWise map
	if len(partnerWiseCount) > limit {
		toRemove := len(partnerWiseCount) - limit
		// remove from lowest priority to highest until we removed required number of partners
		for i := len(updatedPriority) - 1; i >= 0 && toRemove > 0; i-- {
			key := updatedPriority[i]
			if _, exists := partnerWiseCount[key]; exists {
				delete(partnerWiseCount, key)
				toRemove--
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
