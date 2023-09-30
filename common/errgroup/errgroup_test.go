package errgroup

import (
	"fmt"
	"testing"
)

func TestGroup(t *testing.T) {
	var wg Group
	wg.Go(func() error {
		fmt.Println("test ..... 1")
		return nil
	})
	wg.Go(func() error {
		fmt.Println("test ..... 2")
		return nil
	})

	wg.Go(func() error {
		fmt.Println("test ..... 2")
		panic("test")
	})

	err := wg.Wait()
	if err != nil {
		t.Fatal()
	}
}
