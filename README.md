# Property Distribution Algorithm

## Content Outline:
- [Overview](#overview)
- [Inputs and Output](#inputs-and-output)
- [Validation and Assumptions](#validation-and-assumptions)
- [How It Works](#how-it-works)
- [Deterministic Properties](#deterministic-properties)
- [Complexity](#complexity)
- [Running Tests](#running-tests)


## Overview

This project implements a partner slot distribution algorithm that allocates a
fixed number of positions (`limit`) across partners using:

- ratio-based fairness,
- strict priority ordering,
- per-partner capacity constraints (`dbcount`).

The core function is `propertyDistribution(...)`, implemented in Go.

## Inputs and Output

### Inputs

- `ratio map[string]float64`
	- Relative weight for each partner.
- `priority []string`
	- Partners ordered from highest to lowest priority.
- `dbcount map[string]int`
	- Maximum available items for each partner.
- `limit int`
	- Total number of slots to allocate.

### Output

- `[]string`
	- A sequence of partner feeds respecting the priority order.

## Validation and Assumptions

The implementation assumes:

- `limit >= 1`
- `ratio[p] >= 0`
- `dbcount[p] >= 0`

## How It Works

> [!NOTE]
> This algorithm is not optimized yet. It focuses on correctness and clarity, with potential for future performance improvements.

The algorithm runs in stages.

### Step 1. Filter Active Partners

Partners with `dbcount == 0` are removed from consideration while preserving
priority order.

### Step 2. Initial Ratio Allocation (Ceiling)

For each active partner:

$$
allocation[p] = \left( \left\lceil \frac{ratio[p]}{\sum ratio} \cdot limit \right\rceil\right)
$$

Using ceiling prevents small-ratio partners from being rounded down too early
and enforces at least one slot per active partner.

### Step 3. Remove Ceiling Extras

Because ceiling can over-allocate total slots, extra slots are subtracted until
the total equals `limit`.

Strategy used in code:

- first attempt to remove from the highest-priority partner (if that partner has
	more than 1),
- then continue removing from lowest to highest priority,
- never reduce a partner below 1.

### Step 4. Cap by Capacity (`dbcount`)

Any partner whose allocation exceeds `dbcount` is capped. The removed amount is
counted as `freed` slots.

### Step 5. Redistribute Freed Slots

Freed slots are reassigned to partners that still have remaining capacity,
following priority-aware rotation.

### Step 6. Handle Too Many Active Partners

If active partners still exceed `limit` after minimum-1 logic, lowest-priority
partners are dropped until count fits.

### Step 7. Build Final Sequence

The final per-partner counts are converted into an ordered list respecting the
priority order.

## Deterministic Properties

Given the same input maps/slice values, output is deterministic because:

- all operational traversal uses the `priority` slice,
- redistribution follows a fixed queue order.

## Complexity

Let:

- `n` = number of partners
- `L` = `limit`

### Time Complexity

Overall propertyDistribution:
 |  Stages    |   Complexity   |
 | --------- | -------------- |
 | Filter Active Partners | O(n) |
 | Initial Ratio Allocation (Ceiling) | O(n) |
 | Remove Ceiling Extras  | O(n²)  <span style="color: rgb(192, 125, 1);">*needs improvement* |
 | Cap by Capacity | O(n) |
 | Redistribute Freed Slots | O(n + L) |
 | Drop Lowest Priority | O(n) |
 | Build Final Sequence | O(n + L)|

Overall worst-case time complexity:

$$
O(n^2 + L)
$$

### Space Complexity

Auxiliary structures store allocations, capacities, and temporary queues/counts:
`O(n)`.

Returned output sequence stores up to `L` entries: `O(L)`.

Overall space complexity:

$$
O(n + L)
$$

## Running Tests

Use Go test:

```bash
go test ./...
```
