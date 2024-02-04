package main

import (
	"dependor/lib/tokenizer"
	"fmt"
)

func main() {
	tk := tokenizer.New(`const foo = require("foo");`)
	output := tk.TokenizeImports()
	fmt.Println(output[0])
}
