package utils

import "fmt"

func MustInit[T any](fn func() (T, error), name string) T {
	v, err := fn()
	if err != nil {
		panic(fmt.Sprintf("❌ %s init failed: %v", name, err))
	}
	return v
}
