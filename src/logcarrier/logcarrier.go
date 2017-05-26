package main

import "fmt"

func main() {
	config := LoadConfig("/home/emacs/Sources/logcarrier/test.toml")
	fmt.Println(config)
}
