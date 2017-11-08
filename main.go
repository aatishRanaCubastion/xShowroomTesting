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

	//created file
	f := NewFile("main")

	//declaring method
	f.Func().Id("add").Params(
		Id("a").Int(),
		Id("b").Int(),
	).Int().Block(
		Return(Id("a").Op("+").Id("b")),
	)

	//calling method
	f.Func().Id("main").Params().Block(
		Qual("fmt", "Println").Call(Lit("Hello, world")),
		Qual("fmt", "Println").Call(Lit("Aatish Here")),
		Qual("fmt","Println").Call(Id("add(2,3)")),
	)


	fmt.Fprintf(file, "%#v", f)
	fmt.Println("Done")

}
