package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const trendSize = 7

type RowData struct {
	key   []string
	trend []float64
	count int
	data  map[string]interface{}
}

func (d RowData) GetCount() string {
	return strconv.Itoa(d.count)
}

func (d RowData) Get(key string) interface{} {
	return d.data[key]
}

func (d RowData) GetTrend() []float64 {
	return d.trend
}

func (d RowData) GetData() map[string]interface{} {
	return d.data
}

type Store struct {
	sync.RWMutex
	duration time.Duration
	distance int
	keys     []string
	rows     []RowData
}

func NewStore(duration time.Duration, distance int, keys []string) *Store {
	return &Store{
		duration: duration,
		distance: distance,
		keys:     keys,
		rows:     make([]RowData, 0),
	}
}

func (s *Store) SetKeys(keys []string) {
	s.keys = keys
}

func (s *Store) Push(value map[string]interface{}) {
	key := s.Key(value)
	for i := range s.rows {
		if ComputeDistance(key, s.rows[i].key) < s.distance {
			s.rows[i].trend[len(s.rows[i].trend)-1] += 1
			s.rows[i].count++
			s.rows[i].data = value
			return
		}
	}

	data := RowData{
		key:   key,
		trend: make([]float64, trendSize),
		count: 1,
		data:  value,
	}
	data.trend[len(data.trend)-1] += 1
	s.rows = append(s.rows, data)
}

func (s *Store) Len() int {
	return len(s.rows)
}

func (s *Store) Get(i int) RowData {
	if i >= 0 && i < len(s.rows) {
		return s.rows[i]
	}
	return RowData{}
}

func (s *Store) Key(value map[string]interface{}) []string {
	key := make([]string, 0)
	for _, name := range s.keys {
		sub := strings.Split(fmt.Sprintf("%v", value[name]), " ")

		// For short parts of key, double sub length x2.
		// Doubling levenshtein distance for this part of key.
		if len(sub) < s.distance {
			sub = append(sub, sub...)
		}

		key = append(key, sub...)
	}
	return key
}

func (s *Store) Shift() {
	for i := range s.rows {
		t := s.rows[i].trend
		copy(t, t[1:])
		t[len(t)-1] = 0
	}
}
