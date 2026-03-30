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

	exhausted, freed := capAtDBCount(allocated, activeDB)

	if freed > 0 {
		redistributeFreed(allocated, activePriority, activeDB, exhausted, freed)
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
		if count < 1 {
			count = 1
		}
		allocated[key] = count
	}
	return allocated
}

func removeExtras(allocated map[string]int, priority []string, limit int) {
	extras := sumValues(allocated) - limit

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

	// calculate removable
	removable := make(map[string]int, len(allocated))
	for _, key := range priority {
		if allocated[key] > 1 {
			removable[key] = allocated[key] - 1
		}
	}

	var removableList []int
	for _, key := range priority {
		if r, ok := removable[key]; ok {
			removableList = append(removableList, r)
		}
	}

	// sort the slice of pairs by count
	sort.Slice(removableList, func(i, j int) bool {
		return removableList[i] < removableList[j]
	})

	totalRounds := 0
	active := len(removableList) // partners who can still be reduced
	prev := 0

	for i := 0; i < len(removableList); i++ {
		cur := removableList[i] // next exhaustion point
		gap := cur - prev       // how many rounds before this partner exhausts

		cost := gap * active // extras needed for those rounds

		if extras >= cost {
			totalRounds += gap
			extras -= cost
			active-- // this partner is now exhausted
			prev = cur
		} else {
			// cannot finish this gap → partial rounds
			rounds := extras / active
			totalRounds += rounds
			extras -= rounds * active
			break
		}
	}

	for p := range allocated {
		if _, ok := removable[p]; ok {
			reduce := min(removable[p], totalRounds)
			allocated[p] -= reduce
		}
	}

	for i := len(priority) - 1; i >= 0 && extras > 0; i-- {
		p := priority[i]
		allocated[p]--
		extras--
	}
}

// capAtDBCount ensures no partner is allocated more than its db count.
// It returns which partners are now exhausted and the total freed slots.
func capAtDBCount(allocated, activeDB map[string]int) (exhausted map[string]bool, freed int) {
	exhausted = make(map[string]bool)
	for key, count := range allocated {
		if count >= activeDB[key] {
			freed += count - activeDB[key]
			allocated[key] = activeDB[key]
			exhausted[key] = true
		}
	}
	return
}

// redistributeFreed hands freed slots one at a time to non-exhausted partners,
// highest priority first. A partner leaves the rotation once it reaches its
// db count.
func redistributeFreed(
	allocated map[string]int,
	priority []string,
	activeDB map[string]int,
	exhausted map[string]bool,
	freed int,
) {
	// Seed the queue lowest->highest priority.
	queue := make([]string, 0, len(priority))
	for i := len(priority) - 1; i >= 0; i-- {
		key := priority[i]
		if !exhausted[key] {
			queue = append(queue, key)
		}
	}

	for freed > 0 && len(queue) > 0 {
		key := queue[0]
		queue = queue[1:]

		allocated[key]++
		freed--

		if allocated[key] < activeDB[key] {
			queue = append(queue, key) // still has room; re-enqueue
		}
	}
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

// sumValues returns the sum of all values in an int map.
func sumValues(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}
