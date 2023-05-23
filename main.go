package main

import (
	"fmt"

	"github.com/androidsr/sc-go/sorm"
)

func main() {
	builder := sorm.Builder("select * from sys_users")
	builder.Eq("id", 1)
	builder.Like("name", 1)
	builder.Ors(builder.And().Between("age", sorm.BetweenInfo{10, 20}))
	fmt.Println(builder.Build())
}
