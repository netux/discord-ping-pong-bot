package main

import "hash/fnv"

// Hash uses hash/fnv to get the hash of the string s as an int64.
func Hash(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}
