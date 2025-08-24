package main

import (
	"fmt"
	"pnemani1993/gotils/linkedlist"
)

func main() {
	lister := linkedlist.New()
	lister.Insert(1)
	fmt.Print("dropped value: ", lister.RemoveLast(), "\n")
	fmt.Print("dropped value: ", lister.RemoveFirst(), "\n")
	// lister.Insert(2)
	// lister.Insert(3)
	// lister.Insert(4)
	// lister.Insert("something")
	// fmt.Print("Printing the size of the stack: ", lister.Size(), "\n")
	// lister.Print()

	// fmt.Print("dropped value: ", lister.RemoveFirst(), "\n")
	// lister.Print()

	// fmt.Print("dropped value: ", lister.RemoveFirst(), "\n")
	// lister.Print()

	// fmt.Print("dropped value: ", lister.RemoveLast(), "\n")
	// lister.Print()

	// fmt.Print("dropped value: ", lister.RemoveLast(), "\n")
	// lister.Print()

	// fmt.Print("dropped value: ", lister.RemoveLast(), "\n")
	// lister.Print()
}
