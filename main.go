package main

import "fmt"

func main() {
	// This main function is intentionally left empty.
	// The propertyDistribution function is tested via the test files.
	ratio := map[string]float64{
		"11": 50,
		"12": 50,
	}
	priority := []string{"11", "12"}
	dbcount := map[string]int{
		"11": 2,
		"12": 1,
	}
	limit := 5

	result := propertyDistribution(ratio, priority, dbcount, limit)
	fmt.Println(result)
}
