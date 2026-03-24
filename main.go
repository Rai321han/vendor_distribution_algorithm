package main

import (
	"fmt"
	"sort"
)

type pair struct {
	key   string
	value float64
}

func main() {
	// Multiple test cases
	testCases := []map[string]float64{
		{
			"11": 60,
			"12": 25,
			"24": 15,
		},
		{
			"11": 70,
			"12": 15,
			"24": 15,
		},
		{
			"11": 70,
			"24": 30,
		},
		{
			"11": 20,
			"12": 75,
			"24": 5,
		},
		{
			"11": 40,
			"12": 55,
			"24": 10,
		},
		{
			"11": 45,
			"12": 3,
			"14": 7,
			"24": 45,
		},
		{
			"11": 64,
			"14": 6,
			"24": 30,
		},
		{
			"11": 25,
			"12": 35,
			"20": 10,
			"22": 10,
			"24": 15,
			"25": 5,
		},
	}

	limit := 1

	for i, ratio := range testCases {
		fmt.Printf("\n=== Case %d ===\n", i+1)

		// generate priority list based on ratio.
		// more ratio -> higher priority.

		// map → slice
		pairs := []pair{}
		for k, v := range ratio {
			pairs = append(pairs, pair{k, v})
		}

		// sort
		sort.SliceStable(pairs, func(i, j int) bool {
			if pairs[i].value == pairs[j].value {
				return pairs[i].key < pairs[j].key
			}
			return pairs[i].value > pairs[j].value
		})

		// extract priority
		priority := []string{}
		for _, p := range pairs {
			priority = append(priority, p.key)
		}

		// Build DB count
		dbcount := map[string]int{}
		for k := range ratio {
			dbcount[k] = 110
		}

		result := propertyDistribution(ratio, priority, dbcount, limit)

		fmt.Println("Ratio:", ratio)
		fmt.Println("Priority:", priority)
		fmt.Println("DBCount:", dbcount)
		fmt.Println("Result:", result)
		fmt.Printf("Limit -> %d | Result Sequence Length -> %d\n", limit, len(result))
	}
}
