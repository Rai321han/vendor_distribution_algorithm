package main

import (
	"fmt"
	"slices"
	"sort"
	"testing"
)

type pair struct {
	key   string
	value float64
}

type Case struct {
	Ratio    map[string]float64
	Limit    int
	Expected []string
}

func TestPropertyDistributionGivenCases(t *testing.T) {
	testCases := []Case{
		{
			Ratio: map[string]float64{
				"11": 60,
				"12": 25,
				"24": 15,
			},
			Limit:    1,
			Expected: []string{"11"},
		},
		{
			Ratio: map[string]float64{
				"11": 60,
				"12": 25,
				"24": 15,
			},
			Limit:    3,
			Expected: []string{"11", "12", "24"},
		},
		{
			Ratio: map[string]float64{
				"11": 60,
				"12": 25,
				"24": 15,
			},
			Limit:    10,
			Expected: []string{"11", "12", "24", "11", "12", "24", "11", "12", "11", "11"},
		},
		{
			Ratio: map[string]float64{
				"11": 70,
				"12": 15,
				"24": 15,
			},
			Limit:    5,
			Expected: []string{"11", "12", "24", "11", "11"},
		},
		{
			Ratio: map[string]float64{
				"11": 70,
				"24": 30,
			},
			Limit:    1,
			Expected: []string{"11"},
		},
		{
			Ratio: map[string]float64{
				"11": 20,
				"12": 75,
				"24": 5,
			},
			Limit:    1,
			Expected: []string{"12"},
		},
		{
			Ratio: map[string]float64{
				"11": 40,
				"12": 55,
				"24": 10,
			},
			Limit:    1,
			Expected: []string{"12"},
		},
		{
			Ratio: map[string]float64{
				"11": 45,
				"12": 3,
				"14": 7,
				"24": 45,
			},
			Limit:    1,
			Expected: []string{"11"},
		},
		{
			Ratio: map[string]float64{
				"11": 64,
				"14": 6,
				"24": 30,
			},
			Limit:    1,
			Expected: []string{"11"},
		},
		{
			Ratio: map[string]float64{
				"11": 25,
				"12": 35,
				"20": 10,
				"22": 10,
				"24": 15,
				"25": 5,
			},
			Limit:    1,
			Expected: []string{"12"},
		},
	}

	for i, testCase := range testCases {

		pairs := []pair{}
		for k, v := range testCase.Ratio {
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
		for k := range testCase.Ratio {
			dbcount[k] = 110
		}

		t.Run(fmt.Sprintf("Case %d", i+1), func(t *testing.T) {
			got := propertyDistribution(testCase.Ratio, priority, dbcount, testCase.Limit)
			want := testCase.Expected
			if !slices.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}
