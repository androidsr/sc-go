package sno

import (
	"github.com/androidsr/paas-go/syaml"
	"github.com/bwmarrin/snowflake"
)

var (
	Node *snowflake.Node
)

func New(config syaml.SnowflakeInfo) {
	Node, _ = snowflake.NewNode(config.WorkerId)
}

func GetString() string {
	return Node.Generate().String()
}

func GetInt64() int64 {
	return Node.Generate().Int64()
}

func GetBase64() string {
	return Node.Generate().Base64()
}

func GetBase2() string {
	return Node.Generate().Base2()
}
