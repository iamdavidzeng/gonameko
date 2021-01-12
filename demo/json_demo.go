package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Name   string            `json:"name"`
	Gender string            `json:"gender"`
	Args   []string          `json:"args"`
	Kwargs map[string]string `json:"kwargs"`
}

func main() {

	david := Person{
		Name:   "david",
		Gender: "male",
		Args:   []string{"name"},
		Kwargs: map[string]string{"name": "david"},
	}

	out, _ := json.Marshal(david)
	fmt.Println(string(out))
}
