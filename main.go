package main

import (
	"encoding/xml"
	"fmt"

	ts "triple-s/cmd/triple-s"
)

type Plant struct {
	XMLName xml.Name `xml:"plant"`
	Id      int      `xml:"id"`
	Name    string   `xml:"name"`
	Origin  []string `xml:"origin"`
}

func (p Plant) String() string {
	return fmt.Sprintf("Plant id=%v, name=%v, origin=%v",
		p.Id, p.Name, p.Origin)
}

func main() {
	fmt.Println("main was started")
	ts.Run()

	// coffee := &Plant{Id: 27, Name: "Coffee"}
	// coffee.Origin = []string{"Ethiopia", "Brazil"}

	// out, _ := xml.MarshalIndent(coffee, " ", "  ")
	// fmt.Println(out)
	// fmt.Println("123")
	// fmt.Println(string(out))
	// fmt.Println("123")

	// fmt.Println(xml.Header + string(out))

	// var p Plant
	// if err := xml.Unmarshal(out, &p); err != nil {
	// 	panic(err)
	// }
	// // fmt.Println(p)

	// tomato := &Plant{Id: 81, Name: "Tomato"}
	// tomato.Origin = []string{"Mexico", "California"}

	// type Nesting struct {
	// 	XMLName xml.Name `xml:"nesting"`
	// 	Plants  []*Plant `xml:"parent>child>plant"`
	// }

	// nesting := &Nesting{}
	// nesting.Plants = []*Plant{coffee, tomato}

	// out, _ = xml.MarshalIndent(nesting, " ", "  ")
	// fmt.Println(string(out))
}
