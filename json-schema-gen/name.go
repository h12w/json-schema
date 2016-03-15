package main

import (
	"bufio"
	"os"
	"strings"
)

func (m nameMap) exportedGoName(s string) string {
	if name, ok := m[s]; ok {
		s = name
	}
	s = snakeToCamel(s)
	if strings.HasSuffix(s, "Id") {
		return strings.TrimSuffix(s, "Id") + "ID"
	}
	return s
}
func snakeToCamel(s string) string {
	ss := strings.Split(s, "_")
	for i := range ss {
		ss[i] = strings.Title(ss[i])
	}
	return strings.Join(ss, "")
}

type nameMap map[string]string

func readNameMap(filename string) (nameMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	nameMap := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		rec := strings.Split(scanner.Text(), " ")
		if len(rec) < 2 {
			continue
		}
		nameMap[rec[0]] = rec[1]
	}
	return nameMap, scanner.Err()
}
