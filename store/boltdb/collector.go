package boltdb

import (
	"container/list"
	"encoding/json"
	"github.com/boltdb/bolt"
	"time"
)

type document struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
}

type scoreCollector struct {
	k            int
	skip         int
	results      *list.List
	total        uint64
	took         time.Duration
	maxCreatedAt int64
}

func newScorerCollector(k int) *scoreCollector {
	return &scoreCollector{
		k:       k,
		skip:    0,
		results: list.New(),
	}
}

func newScorerSkipCollector(k, skip int) *scoreCollector {
	return &scoreCollector{
		k:       k,
		skip:    skip,
		results: list.New(),
	}
}

func (sc *scoreCollector) Total() uint64 {
	return sc.total
}

func (sc *scoreCollector) Collect(cur *bolt.Cursor) error {
	startTime := time.Now()
	var doc document
	for k, v := cur.First(); k != nil; k, v = cur.Next() {
		json.Unmarshal(v, &doc)
		sc.collectSingle(doc)
	}
	// compute search duration
	sc.took = time.Since(startTime)
	return nil
}

func (sc *scoreCollector) collectSingle(doc document) {
	// increment total hits
	sc.total++

	// update max create time
	if doc.CreatedAt > sc.maxCreatedAt {
		sc.maxCreatedAt = doc.CreatedAt
	}

	for e := sc.results.Front(); e != nil; e = e.Next() {
		curr := e.Value.(document)
		if doc.CreatedAt < curr.CreatedAt {

			sc.results.InsertBefore(doc, e)
			// if we just made the list too long
			if sc.results.Len() > (sc.k + sc.skip) {
				// remove the head
				sc.results.Remove(sc.results.Front())
			}
			return
		}
	}
	// if we got to the end, we still have to add it
	sc.results.PushBack(doc)
	if sc.results.Len() > (sc.k + sc.skip) {
		// remove the head
		sc.results.Remove(sc.results.Front())
	}
}

func (sc *scoreCollector) Results() []string {
	if sc.results.Len()-sc.skip > 0 {
		rv := make([]string, sc.results.Len()-sc.skip)
		i := 0
		skipped := 0
		for e := sc.results.Back(); e != nil; e = e.Prev() {
			if skipped < sc.skip {
				skipped++
				continue
			}
			rv[i] = e.Value.(document).ID
			i++
		}
		return rv
	}
	return nil
}
