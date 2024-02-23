package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"unicode/utf8"
)

var (
	textLength  = flag.Int("l", 1000, "length of text to generate")
	ngramLength = flag.Int("n", 3, "ngram length")
)

func main() {
	flag.Parse()
	var (
		stat      ngramStat = ngramStat{ngrams: make(map[string]*ngramVariants)}
		ngram     string
		prevNgram string
	)

	for i, fname := range flag.Args() {
		fmt.Printf("Processing file %d/%d\r", i, len(flag.Args()))
		f, err := os.Open(fname)
		if err != nil {
			log.Panic("opening file:", err)
		}
		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanRunes)
		for scanner.Scan() {
			nextRune := scanner.Text()
			if utf8.RuneCountInString(ngram) < *ngramLength {
				ngram += nextRune
				continue
			}
			_, n := utf8.DecodeRuneInString(ngram)
			ngram = ngram[n:] + nextRune
			if prevNgram != "" {
				stat.add(prevNgram, ngram)
			}

			prevNgram = ngram
		}
	}
	fmt.Printf("\nGot %d ngrams\n", len(stat.ngrams))
	fmt.Println(generate(stat, *textLength))
	//dumpNgrams(stat)

}
