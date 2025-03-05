//go:build !solution

package hotelbusiness

import "sort"

type Guest struct {
	CheckInDate  int
	CheckOutDate int
}

type Load struct {
	StartDate  int
	GuestCount int
}

func ComputeLoad(guests []Guest) []Load {
	datesToLoad := make(map[int]int)

	for _, guest := range guests {
		datesToLoad[guest.CheckInDate]++
		datesToLoad[guest.CheckOutDate]--
	}

	dates := make([]int, 0, len(datesToLoad))
	for date := range datesToLoad {
		dates = append(dates, date)
	}
	sort.Ints(dates)

	load := []Load{}
	loadness := 0
	lastLoadness := -1
	for _, date := range dates {
		loadness += datesToLoad[date]
		if lastLoadness != loadness {
			load = append(load, Load{date, loadness})
			lastLoadness = loadness
		}
	}

	return load
}
