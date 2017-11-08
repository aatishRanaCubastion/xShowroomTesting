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
		Qual("fmt","Println").Call(Lit("Aatish Here")),
	)

	//declaring method
	f.Func().Id("add").Params(
		Id("a").Int(),
	).Block()


	fmt.Fprintf(file, "%#v", f)
	fmt.Println("Done")

}
