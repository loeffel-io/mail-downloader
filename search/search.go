package search

import (
	"github.com/loeffel-io/tax/counter"
	"strings"
)

type Search struct {
	Search []string
	Data   string
}

func (search *Search) Find() bool {
	for _, row := range search.Search {
		count := counter.CreateCounter()
		split := strings.Split(row, ",")

		for _, cell := range split {
			if strings.Contains(search.Data, strings.TrimSpace(strings.ToLower(cell))) {
				count.Increase()
			}
		}

		if count.Current() == len(split) {
			return true
		}
	}

	return false
}
