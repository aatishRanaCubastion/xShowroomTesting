package main

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"os"
	"log"
)

func main() {

	file, err := os.Create("xShowroom.go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()


	f := NewFile("main")
	f.Func().Id("main").Params().Block(
		Qual("fmt", "Println").Call(Lit("Hello, world")),
	)
	fmt.Fprintf(file, "%#v", f)
	fmt.Println("Done")
}
