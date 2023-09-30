package util

import (
	"fmt"
	"testing"
)

func TestAES(t *testing.T) {
	key := []byte(fmt.Sprintf("%16b", 12389))
	txt := "this is a test example"
	result, err := AESEncrypt(txt, key)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result)

	s := "RPo_LUf38_OwlWGqirz1YYUIc9SYYkE2yjmgQy1o5Z6K9q4hZC383VzxCZz-bMoh"

	expect, err := AESDecrypt(s, key)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("decrypt: [%s]\n", expect)
	if expect != txt {
		t.Fatal()
	}

	expect, err = AESDecrypt(result, key)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("decrypt: [%s]\n", expect)
	if expect != txt {
		t.Fatal()
	}
}
