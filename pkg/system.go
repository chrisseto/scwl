package pkg

import (
	"context"
	"fmt"
	"math/rand"
)

type Command interface {
}

type StateNode interface {
	Children() []StateNode
}

type System interface {
	Execute(context.Context, Command) error

	State(context.Context) (StateNode, error)
}

func RandomString() string {
	words := []string{
		"big",
		"small",
		"llama",
		"bat",
		"horse",
		"Abaris",
		"abarthrosis",
		"abarticular",
		"abarticulation",
		"abas",
		"abase",
		"abased",
		"abasedly",
		"abasedness",
		"abasement",
		"abaser",
		"Abasgi",
		"abash",
		"abashed",
		"abashedly",
		"abashedness",
		"abashless",
		"abashlessly",
		"abashment",
		"abasia",
		"abasic",
		"abask",
		"Abassin",
		"abastardize",
		"abatable",
		"abate",
		"abatement",
		"abater",
		"abatis",
		"abatised",
		"abaton",
		"abator",
		"abattoir",
		"Abatua",
		"abature",
		"abave",
		"abaxial",
		"abaxile",
		"abaze",
		"abb",
		"Abba",
		"abbacomes",
		"abbacy",
		"Abbadide",
	}

	return fmt.Sprintf(
		"%s_%s",
		words[rand.Intn(len(words))],
		words[rand.Intn(len(words))],
	)
}
