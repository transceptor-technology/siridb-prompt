package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type query struct {
	req string
	res interface{}
	err error
}

func newQuery(req string) *query {
	q := query{
		req: req,
		res: nil,
		err: nil,
	}
	return &q
}

func readJSON(b []byte, v *interface{}) error {
	reader := bytes.NewReader(b)
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(v); err != nil {
		return err
	}
	return nil
}

func readCSV(b []byte, v *interface{}) error {
	reader := bytes.NewReader(b)
	res, err := parseCsv(reader)
	if err != nil {
		return err
	}

	*v = res
	return nil
}

func importFromFile(fn string, timeout uint16) (interface{}, error) {
	var v interface{}
	var loader func(b []byte, v *interface{}) error

	if strings.HasSuffix(strings.ToLower(fn), ".json") {
		loader = readJSON
	} else if strings.HasSuffix(strings.ToLower(fn), ".csv") {
		loader = readCSV
	} else {
		return nil, fmt.Errorf("only .json or .csv files are supported")
	}

	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	err = loader(data, &v)
	if err != nil {
		return nil, err
	}

	return client.Insert(v, timeout)
}

func (q *query) dumpToFile(fn string) error {
	var s string
	var e error
	if strings.HasSuffix(strings.ToLower(fn), ".json") {
		s, e = q.json()
	} else if strings.HasSuffix(strings.ToLower(fn), ".csv") {
		s, e = q.csv()
	} else {
		return fmt.Errorf("only .json or .csv files are supported")
	}

	if e != nil {
		return e
	}

	e = ioutil.WriteFile(fn, []byte(s), 0644)
	if e != nil {
		return e
	}
	return fmt.Errorf("last result successfully dumped to: %s", fn)
}

func (q *query) parse(timeout uint16) {
	if strings.HasPrefix(q.req, "import ") {
		fn := strings.TrimSpace(q.req[6:])
		q.res, q.err = importFromFile(fn, timeout)
	} else {
		q.res, q.err = client.Query(q.req, timeout)
	}
}

func (q *query) json() (string, error) {
	var b []byte
	var err error
	if q.res == nil {
		return "", fmt.Errorf("nothing to JSONify")
	}
	if b, err = json.Marshal(q.res); err != nil {
		return "", err
	}
	return string(b), nil
}

func (q *query) csv() (string, error) {
	if q.res == nil {
		return "", fmt.Errorf("nothing to CSVify")
	}
	m, ok := q.res.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("cannot convert data to CSV")
	}
	return toCsv(m)
}
