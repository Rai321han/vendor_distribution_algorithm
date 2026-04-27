package main

import (
	"maps"
	"math"
	"sort"
)

// propertyDistribution distributes a limited number of slots among partners
// based on their specified ratios and priorities, capped by each partner's
// available db count.
//
// Parameters:
//   - ratio:    partners → share weight
//   - priority: partners ordered highest → lowest priority
//   - dbcount:  partners → available counts in the database
//   - limit:    total slots
//
// Returns a sequence of partners' feeds ordered by priority.
//
// Algorithm:
//  1. Filter out partners with dbcount == 0 and normalise ratios.
//  2. Ceiling-allocate each partner's initial share from `limit`.
//  3. Remove over-allocation (extras) from ceiling rounding:
//     subtract from the highest-priority partner first (if > 1),
//     then sweep lowest→highest until extras reach zero.
//  4. Cap each partner at its dbcount; collect freed slots.
//  5. Redistribute freed slots to non-exhausted partners, highest priority first.
//  6. If active partners > limit (all at count 1), drop the lowest-priority ones.
//  7. Emit the final sequence via round-robin in priority order.
//
// Constraints:
//
//	limit >= 1
//	dbcount[p] >= 0, ratio[p] >= 0 for all partners p
func propertyDistribution(
	ratio map[string]float64,
	priority []string,
	dbcount map[string]int,
	limit int,
) []string {
	if len(ratio) != len(priority) || len(ratio) != len(dbcount) {
		return []string{}
	}

	activePriority, activeRatio, activeDB := filterActive(priority, ratio, dbcount)
	if len(activePriority) == 0 {
		return []string{}
	}

	allocated := ceilAllocate(activePriority, activeRatio, limit)
	removeExtras(allocated, activePriority, limit)

	freed := capAtDBCount(allocated, activeDB)

	if freed > 0 {
		redistributeFreed(allocated, activePriority, activeDB, freed)
	}

	if len(allocated) > limit {
		dropLowestPriority(allocated, activePriority, len(allocated)-limit)
	}

	return buildSequence(allocated, activePriority)
}

// filterActive returns only partners whose dbcount > 0, preserving order.
func filterActive(
	priority []string,
	ratio map[string]float64,
	dbcount map[string]int,
) (activePriority []string, activeRatio map[string]float64, activeDB map[string]int) {
	activeRatio = make(map[string]float64, len(priority))
	activeDB = make(map[string]int, len(priority))

	for _, key := range priority {
		if dbcount[key] == 0 {
			continue
		}
		activePriority = append(activePriority, key)
		activeRatio[key] = ratio[key]
		activeDB[key] = dbcount[key]
	}
	return
}

// ceilAllocate assigns each partner at least 1 slot using ceiling division so
// that the sum of initial allocations is >= limit.
func ceilAllocate(priority []string, activeRatio map[string]float64, limit int) map[string]int {
	var totalRatio float64
	for _, key := range priority {
		totalRatio += activeRatio[key]
	}

	allocated := make(map[string]int, len(priority))
	for _, key := range priority {
		share := activeRatio[key] / totalRatio
		count := int(math.Ceil(share * float64(limit)))
		count = max(count, 1)
		allocated[key] = count
	}
	return allocated
}

// removeExtras removes any overallocation from ceiling rounding, starting with the highest-priority partner (if it has more than 1 slot),
// then sweeping from lowest to highest priority until we've removed enough extras.
// This ensures we don't unfairly penalise lower-priority partners for the ceiling rounding of higher-priority ones,
// while still respecting the priority order when removing extras.
func removeExtras(allocated map[string]int, priority []string, limit int) {
	extras := 0
	for _, v := range allocated {
		extras += v
	}
	extras -= limit

	if extras == 0 {
		return
	}

	for _, key := range priority {
		if allocated[key] > 1 {
			allocated[key]--
			extras--
			break
		}
	}

	if extras == 0 {
		return
	}

	capacity := buildCapacity(priority, func(k string) int {
		return allocated[k] - 1
	})

	rebalanceAllocation(allocated, priority, capacity, extras, -1)
}

// redistributeFreed takes any slots freed by capping at db count and redistributes them to non-exhausted partners in priority order,
// respecting their capacity to receive more slots (activeDB - allocated). It uses the same rebalanceAllocation helper as removeExtras,
// but with positive units to add rather than remove.
func redistributeFreed(
	allocated map[string]int,
	priority []string,
	activeDB map[string]int,
	freed int,
) {

	if freed == 0 {
		return
	}
	capacity := buildCapacity(priority, func(k string) int {
		return activeDB[k] - allocated[k]
	})
	rebalanceAllocation(allocated, priority, capacity, freed, +1)
}

// buildCapacity constructs a map of partner → capacity based on the provided condition function.
func buildCapacity(priority []string, cond func(string) int) map[string]int {
	c := map[string]int{}
	for _, k := range priority {
		if v := cond(k); v > 0 {
			c[k] = v
		}
	}
	return c
}

// capAtDBCount ensures no partner is allocated more than its db count.
// It returns the total freed slots.
func capAtDBCount(allocated, activeDB map[string]int) int {
	freed := 0
	for key, count := range allocated {
		if count >= activeDB[key] {
			freed += count - activeDB[key]
			allocated[key] = activeDB[key]
		}
	}
	return freed
}

// dropLowestPriority removes n lowest-priority partners from the allocation
// map. Used only when every active partner has exactly 1 slot but there are
// more partners than the limit allows.
func dropLowestPriority(allocated map[string]int, priority []string, n int) {
	for i := len(priority) - 1; i >= 0 && n > 0; i-- {
		if _, ok := allocated[priority[i]]; ok {
			delete(allocated, priority[i])
			n--
		}
	}
}

// rebalanceAllocation is a helper for both removeExtras and redistributeFreed.
//
// It takes:
//
//	participants:
//
// a list of integers indicating availiability for removal or addition.
// For removeExtras, this is how many slots can be removed from each partner (allocated - 1).
// For redistributeFreed, this is how many slots can be added to each partner (activeDB - allocated).
//
//	priority:
//
// the list of partners in priority order.
//
//	allocated:
//
// the current allocation map that we will modify in place.
//
//	capacity:
//
// the maximum capacity for each partner (either removable or addable).
//
//	units:
//
// how many slots we need to remove (negative) or add (positive).
//
//	sign:
//
// +1 indicates we are adding slots (redistributeFreed), -1 indicates we are removing slots (removeExtras).
//
// This algorithm is more efficient than removing/adding one slot at a time.
// It calculates how many full rounds we can do with the current participants before we exhaust the next partner, and removes/adds that many slots in one go, then moves on to the next partner.
// This way we can remove/add multiple slots in one iteration instead of one by one, which is more efficient when we have a large number of slots to remove/add.
// If we exhaust all participants but still have slots to remove/add, we will continue removing/adding from the lowest priority partners until we've removed/added all required slots respecting the capacity constraints.
func rebalanceAllocation(allocated map[string]int,
	priority []string,
	capacity map[string]int,
	units int,
	sign int,
) {

	type pair struct {
		key string
		cap int
	}
	var parts []pair
	for k, c := range capacity {
		parts = append(parts, pair{k, c})
	}

	sort.Slice(parts, func(i, j int) bool { return parts[i].cap < parts[j].cap })

	totalRounds := 0
	active := len(parts) // participants who can still be reduced
	prev := 0

	for i := range parts {
		cur := parts[i].cap // next exhaustion point
		gap := cur - prev   // how many rounds before this participant exhausts

		cost := gap * active // needs for those rounds

		if units >= cost {
			totalRounds += gap
			units -= cost
			active-- // this participant is now exhausted
			prev = cur
		} else {
			rounds := units / active
			totalRounds += rounds
			units -= rounds * active
			break
		}
	}

	for p := range allocated {
		if cap, ok := capacity[p]; ok {
			change := min(cap, totalRounds) // how many we can actually remove/add for this partner
			allocated[p] += change * sign   // apply the change with the correct sign
		}
	}

	// If we still have units to remove/add after exhausting all participants, we continue removing/adding from the lowest priority partners until we've removed/added all required slots respecting the capacity constraints.
	for i := len(priority) - 1; i >= 0 && units > 0; i-- {
		p := priority[i]
		if _, ok := capacity[p]; ok {
			allocated[p] += sign
			units--
		}
	}
}

// buildSequence emits all allocated partner slots in round-robin priority
// order (highest priority partner first in each round).
func buildSequence(allocated map[string]int, priority []string) []string {
	active := make([]string, 0, len(allocated))
	for _, key := range priority {
		if allocated[key] > 0 {
			active = append(active, key)
		}
	}

	counts := make(map[string]int, len(allocated))
	maps.Copy(counts, allocated)

	var result []string
	for len(active) > 0 {
		next := active[:0]
		for _, key := range active {
			result = append(result, key)
			counts[key]--
			if counts[key] > 0 {
				next = append(next, key)
			}
		}
		active = next
	}
	return result
}
