package types_test

import (
	"encoding/json/v2"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.yorun.ai/skel/types"
)

func Example() {
	payload := struct {
		Price types.Decimal   `json:"price"`
		At    types.Timestamp `json:"at"`
		File  types.Binary    `json:"file"`
	}{
		Price: types.NewDecimal(decimal.RequireFromString("1.00")),
		At:    types.NewTimestamp(time.Date(2026, time.October, 6, 12, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))),
		File:  types.Binary("hello"),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(encoded))
	// Output: {"price":"1.00","at":"2026-10-06T04:00:00Z","file":"aGVsbG8="}
}
