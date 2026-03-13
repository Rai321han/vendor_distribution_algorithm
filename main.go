package main

import (
	"math"
)

type Ratio struct {
	Key   string
	Value float64
}

// After rounding
//1.1 If +1 -> subtract from highest priority
//1.2 If +2 or mote -> subtract 1 from highest priority, 1 from from lowest priority, then lower -> goes on

// DB check
//2.1 If missing, addition -> lower priority to higher priority

// Solution
// Rounding -> use ceil -> always sum >= limit -> rule 1.1 or rule 1.2

func vda(ratio map[string]float64, priority []string, dbcount map[string]int, limit int) {

	// Step: DB count check

	var availableRatio float64
	var ratioSlice []Ratio

	// Remove that shares that has 0 db count, calculate based on remaining sum shares
	for key, vendor := range dbcount {
		if vendor == 0 {
			delete(dbcount, key)
			delete(ratio, key)
			continue
		}
		availableRatio += ratio[key]
		ratioSlice = append(ratioSlice, Ratio{
			Key:   key,
			Value: ratio[key],
		})
	}

	// Step: Vendor Wise Initial Count

	vendorCount := make(map[string]int)

	// No vendors with non-zero db counts and non-zero ratios can have 0 share
	reminders := limit - len(ratio)

	// Calculate the initial counts from ratio and limit
	for key, value := range ratio {
		vendorCount[key] = int(math.Ceil((value / availableRatio) * float64(reminders)))
	}

	// sum of initial counts
	vendorCountTotal := 0
	for _, count := range vendorCount {
		vendorCountTotal += count
	}

	// calculate new reminders
	// new_reminders := reminders - vendorCountTotal

	// sort desc based on ratio
	// slices.SortFunc(ratioSlice, func(a, b Ratio) int {
	// 	if a.Value > b.Value {
	// 		return -1
	// 	}
	// 	if a.Value < b.Value {
	// 		return 1
	// 	}
	// 	return 0
	// })

	// // subtract extras

	// for _, ratio := range ratioSlice {
	// 	if new_reminders == 0 {
	// 		break
	// 	}
	// 	vendorCount[ratio.Key]++
	// 	new_reminders--
	// }

	// // add +1 to all vendors which was deducted in the beginning for making space for 1 share for each vendor
	// for key := range vendorCount {
	// 	vendorCount[key]++
	// }

	// // Calculate total spots remains and total requires
	// slotRemains := 0
	// allocated := 0
	// for key, count := range vendorCount {
	// 	if dbcount[key] < count {
	// 		slotRemains += count - dbcount[key]
	// 		vendorCount[key] = dbcount[key]
	// 		allocated++
	// 	}
	// }

	// // if no slots left, then return sequence based on priority and vendor count
	// if slotRemains == 0 {
	// 	// return result(vendorCount, priority)
	// }

	// // distribution from lower priority to higher priority until slots remains or all vendors are fulfilled
	// for slotRemains > 0 && allocated < len(vendorCount) {
	// 	//
	// }
	// // return result(vendors, priorityList)
}

func main() {
	ratio := map[string]float64{
		"12": 30,
		"24": 20,
		"11": 50,
	}

	priority := []string{"11", "12", "24"}

	dbCount := map[string]int{
		"11": 100,
		"12": 200,
		"24": 300,
	}

	limit := 4

	vda(ratio, priority, dbCount, limit)
}
