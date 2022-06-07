package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Index struct {
	row int
	col int
}

func (index Index) ToString() string {
	return fmt.Sprintf("%d%d", index.row, index.col)
}

func ParseIndex(strIndex string) Index {
	row, _ := strconv.Atoi(strIndex[0:1])
	col, _ := strconv.Atoi(strIndex[1:2])
	return Index{row, col}
}

type PolyKey struct {
	alphabet []rune
	nrows    int
	ncols    int
}

func (polyKey *PolyKey) CharToIndex(char rune) (Index, error) {
	row_i := 0

	for alph_i, alph_char := range polyKey.alphabet {

		if alph_i%polyKey.ncols == 0 && alph_i > 0 {
			row_i++
		}

		if alph_char == char {
			return Index{row_i, alph_i % polyKey.ncols}, nil
		}
	}

	return Index{-1, -1}, errors.New("Character out of range")
}

func (polyKey *PolyKey) IndexToChar(index Index) rune {
	flatIndex := (index.row * polyKey.ncols) + index.col

	return polyKey.alphabet[flatIndex]
}

func ReadKey(fileName string) (*PolyKey, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	nrows := len(records)
	ncols := len(records[0])
	alphLen := nrows * ncols
	alphabet := make([]rune, alphLen, alphLen)

	for row_i, row := range records {
		for col_i, item := range row {
			var first rune
			for _, c := range item {
				first = c
				break
			}
			alphabet[(row_i*ncols)+col_i] = first
		}
	}

	return &PolyKey{alphabet, nrows, ncols}, nil
}

func EncodeStream(reader *bufio.Reader, writer *bufio.Writer, polyKey *PolyKey) error {
	defer writer.Flush()

	for {
		rune, _, error := reader.ReadRune()
		if error == io.EOF {
			break
		}

		index, error := polyKey.CharToIndex(rune)
		if error != nil {
			return error
		}

		writer.WriteString(index.ToString())
	}
	return nil
}

func DecodeStream(reader *bufio.Reader, writer *bufio.Writer, polyKey *PolyKey) error {
	defer writer.Flush()
	counter := 0
	var buffer rune

	for {
		rune, _, error := reader.ReadRune()
		if error == io.EOF {
			break
		}

		if counter == 0 {
			buffer = rune
			counter += 1
			continue
		}

		counter = 0
		row, _ := strconv.Atoi(string(buffer))
		col, _ := strconv.Atoi(string(rune))
		index := Index{row, col}
		writer.WriteRune(polyKey.IndexToChar(index))
	}
	return nil
}

func main() {
	keyFilePtr := flag.String("key", "key.csv", "key file")
	decodePtr := flag.Bool("decode", false, "to decode")

	flag.Parse()

	polyKey, err := ReadKey(*keyFilePtr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	if *decodePtr {
		DecodeStream(reader, writer, polyKey)
	} else {
		EncodeStream(reader, writer, polyKey)
	}
}
