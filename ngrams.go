package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"unicode/utf8"
)

type ngramVariants struct {
	weights map[string]int
	weight  int
}

type ngramStat struct {
	ngrams map[string]*ngramVariants
}

func loadStat(db *sql.DB) (ngramStat, error) {
	stat := ngramStat{
		ngrams: make(map[string]*ngramVariants),
	}
	rows, err := db.Query(`SELECT ngram, nextNgram, weight FROM ngrams`)
	if err != nil {
		return stat, fmt.Errorf("querying db: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ngram, nextNgram string
		var weight int
		if err := rows.Scan(&ngram, &nextNgram, &weight); err != nil {
			return stat, fmt.Errorf("scanning the result: %w", err)
		}
		if v, ok := stat.ngrams[ngram]; !ok {
			stat.ngrams[ngram] = &ngramVariants{
				weights: map[string]int{nextNgram: weight},
				weight:  weight,
			}
		} else {
			v.weights[nextNgram] = weight
			v.weight += weight
		}
	}
	if err := rows.Err(); err != nil {
		return stat, fmt.Errorf("scanning the result: %w", err)
	}
	return stat, nil
}

func (v ngramVariants) get() string {
	n := rand.Intn(v.weight)
	for k, v := range v.weights {
		n -= v
		if n < 0 {
			return k
		}
	}
	panic("invalid weights?")
}

func (n ngramStat) getRandom() string {
	for k := range n.ngrams {
		return k
	}
	return ""
}

func (s *ngramStat) add(ngram, nextNgram string) {
	if v, ok := s.ngrams[ngram]; !ok {
		s.ngrams[ngram] = &ngramVariants{
			weights: map[string]int{nextNgram: 1},
			weight:  1,
		}
		return
	} else {
		v.weight++
		v.weights[nextNgram]++
	}
}

func (s ngramStat) getNext(ngram string) string {
	if options := s.ngrams[ngram]; options == nil {
		return s.getRandom()
	} else {
		return options.get()
	}
}

func (s ngramStat) generate(length int) string {
	var (
		current = s.getRandom()
		result  = current
		next    string
	)
	for i := 0; i < length; i++ {
		next = s.getNext(current)
		c, _ := utf8.DecodeLastRuneInString(next)
		result += string(c)
		current = next
	}

	return result
}

func (s ngramStat) save(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Begin(): %w", err)
	}
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO ngrams(ngram, nextNgram, weight) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("Prepare(): %w", err)
	}
	defer stmt.Close()
	for ngram, variants := range s.ngrams {
		for nextNgram, weight := range variants.weights {
			if _, err := stmt.Exec(ngram, nextNgram, weight); err != nil {
				return fmt.Errorf("inserting into db: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Commit(): %w", err)
	}
	return nil
}

func dumpNgrams(stat ngramStat) {
	for k, v := range stat.ngrams {
		fmt.Printf("%s:\n", k)
		for n, c := range v.weights {
			fmt.Printf("\t%s -> %d\n", n, c)
		}
		fmt.Println()
	}
	fmt.Printf("%d ngrams collected\n", len(stat.ngrams))
}
