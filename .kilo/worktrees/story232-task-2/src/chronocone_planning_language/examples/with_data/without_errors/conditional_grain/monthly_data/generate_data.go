package main

import (
	"fmt"
	"os"
	"strconv"

	cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

func makeDayRange(start, end int) []string {
	result := make([]string, 0, end-start+1)
	for i := start; i <= end; i++ {
		result = append(result, strconv.Itoa(i))
	}
	return result
}

func main() {
	allMonths := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
	daysInMonth := map[string]int{
		"January": 31, "February": 28, "March": 31,
		"April": 30, "May": 31, "June": 30,
		"July": 31, "August": 31, "September": 30,
		"October": 31, "November": 30, "December": 31,
	}

	dimensions := map[string][]string{
		"Month":      allMonths,
		"DayInMonth": makeDayRange(1, 31), // full range; conditionals restrict per month
	}

	grain := []cplio.GrainMeasure{
		{Name: "Monthly", Values: allMonths},
	}

	// Build conditional mappings: DayInMonth values depend on Month.
	// Monthly also depends on Month so that each row's Monthly matches its Month.
	dayInMonthValues := make(map[string][]string)
	monthlyValues := make(map[string][]string)
	for _, m := range allMonths {
		dayInMonthValues[m] = makeDayRange(1, daysInMonth[m])
		monthlyValues[m] = []string{m}
	}
	cond := map[string]cplio.ConditionalDimMapping{
		"DayInMonth": {
			Parent: "Month",
			Values: dayInMonthValues,
		},
		"Monthly": {
			Parent: "Month",
			Values: monthlyValues,
		},
	}

	err := cplio.GenerateParquetFileWithConditionals("monthly_data.parquet", dimensions, grain, nil, cond)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated monthly_data.parquet")
}
