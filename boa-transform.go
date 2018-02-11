package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

var recordLexer = lexer.Must(lexer.Regexp(
	`(?m)` +
		`(\s+)` +
		`|(?P<Date>[0-9]{2}\/[0-9]{2})` +
		`|(?P<Float>\d+(?:\.\d+)?)` +
		`|(?P<Description>.+) {64}` +
		`|(?P<Stopword>Virtual Card)` +
		`|(?P<Whitespace>\s+)`,
))

type Record struct {
	TransactionDate string ` @Date`
	PostingDate     string ` @Date `
	Description     string ` @Description `
	ReferenceNumber string ` @Float `
	AccountNumber   string `( @Stopword | @Float ) `
	Amount          string ` @Float`
}

func fixupRecord(rec *Record) {
	rec.Description = strings.TrimSpace(rec.Description)
	year := fmt.Sprintf("%v", time.Now().Year())
	rec.TransactionDate = year + "/" + rec.TransactionDate
	rec.Amount = "-" + rec.Amount
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func convertInputToRecords(r io.Reader) {
	parser, err := participle.Build(&Record{}, recordLexer)
	check(err)
	rec := &Record{}

	scanner := bufio.NewScanner(r)
	fmt.Println("date,description,amount")
	for scanner.Scan() {
		str := scanner.Text()
		match, err := regexp.Match(`^\s+[0-9]{2}/[0-9]{2}`, []byte(str))
		check(err)
		if match {
			err := parser.ParseString(str, rec)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: Failed to parse record, skipping")
				fmt.Fprintln(os.Stderr, str)
			} else {
				fixupRecord(rec)
				fmt.Printf(`"%s","%s","%s"`+"\n", rec.TransactionDate, rec.Description, rec.Amount)

			}
		}
	}

}

func main() {
	flag.Parse()

	switch flag.NArg() {
	case 0:
		convertInputToRecords(os.Stdin)
	case 1:
		f, _ := os.Open(flag.Arg(0))
		convertInputToRecords(bufio.NewReader(f))
	}

}
