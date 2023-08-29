package pkg

import (
	"context"
	"math/rand"
	"strings"
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

func RandomString(prefixes ...string) string {
	for len(prefixes) < 3 {
		prefixes = append(prefixes, words[rand.Intn(len(words))])
	}

	return strings.Join(prefixes, "_")
}

// Generated with: cat /usr/share/dict/words | rg -so '[a-z]{4,6}' | head -n 70
var words = []string{
	"aalii",
	"aardva",
	"aardwo",
	"aron",
	"aronic",
	"aronic",
	"aronit",
	"aronit",
	"babdeh",
	"babua",
	"abac",
	"abaca",
	"abacat",
	"abacay",
	"abacin",
	"abacin",
	"ation",
	"abacis",
	"abacis",
	"aback",
	"abacti",
	"abacti",
	"nally",
	"abacti",
	"abacto",
	"abacul",
	"abacus",
	"badite",
	"abaff",
	"abaft",
	"abaisa",
	"abaise",
	"abaiss",
	"abalie",
	"nate",
	"abalie",
	"nation",
	"abalon",
	"bama",
	"abampe",
	"abando",
	"abando",
	"nable",
	"abando",
	"abando",
	"nedly",
	"abando",
	"abando",
	"abando",
	"nment",
	"banic",
	"bantes",
	"abapti",
	"ston",
	"baramb",
	"baris",
	"abarth",
	"rosis",
	"abarti",
	"cular",
	"abarti",
	"culati",
	"abas",
	"abase",
	"abased",
	"abased",
	"abased",
	"ness",
	"abasem",
	"abaser",
}