package main

import (
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

func (v ngramVariants) get() string {
	n := rand.Intn(v.weight)
	for k, v := range v.weights {
		n -= v
		if n < 0 {
			return k
		}
	}
	return "ERROR"
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

func generate(stat ngramStat, length int) string {
	var (
		current = stat.getRandom()
		result  = current
		next    string
	)
	for i := 0; i < length; i++ {
		next = stat.getNext(current)
		c, _ := utf8.DecodeLastRuneInString(next)
		result += string(c)
		current = next
	}

	return result
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
