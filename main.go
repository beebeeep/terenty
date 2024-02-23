package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"
)

const (
	_ngramTableSQL = `CREATE TABLE IF NOT EXISTS ngrams(
		ngram text,
		nextNgram text,
		weight integer,
		PRIMARY KEY (ngram, nextNgram))`
)

var (
	textLength  = flag.Int("l", 1000, "length of text to generate")
	ngramLength = flag.Int("n", 3, "ngram length")
	dbFile      = flag.String("db", "ngrams.db", "DB file")
	mode        = flag.String("m", "text", "mode (text, read)")
)

func main() {
	flag.Parse()
	switch *mode {
	case "read":
		db, err := initDb()
		if err != nil {
			log.Fatal(err)
		}
		if err := generateNgrams(db); err != nil {
			log.Fatal(err)
		}
	case "text":
		db, err := initDb()
		if err != nil {
			log.Fatal(err)
		}
		text, err := generateText(db, *textLength)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(text)
	}
}

func initDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", *dbFile)
	if err != nil {
		return nil, fmt.Errorf("opening the db: %w", err)
	}
	if _, err := db.Exec(_ngramTableSQL); err != nil {
		return nil, fmt.Errorf("creating schema: %w", err)
	}
	return db, nil
}

func generateNgrams(db *sql.DB) error {
	var (
		ngram     string
		prevNgram string
	)
	stat, err := loadStat(db)
	if err != nil {
		return fmt.Errorf("loading existing stats: %w", err)
	}
	fmt.Printf("Loaded %d ngrams\n", len(stat.ngrams))
	files := getFilesToRead()
	for i, fname := range files {
		fmt.Printf("Processing file %d/%d\r", i, len(files))
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
	fmt.Printf("\nRead %d ngrams\n", len(stat.ngrams))
	return stat.save(db)
}

func generateText(db *sql.DB, length int) (string, error) {
	stat, err := loadStat(db)
	if err != nil {
		return "", fmt.Errorf("loading statistics: %w", err)
	}
	return stat.generate(length), nil
}

func getFilesToRead() []string {
	if len(flag.Args()) > 0 {
		return flag.Args()
	}
	files := make([]string, 0, 10)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}
	return files
}
