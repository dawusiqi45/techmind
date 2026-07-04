package snowflake

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

// Init 初始化雪花 ID 生成器，machineID 范围 0-1023
func Init(machineID int64) error {
	n, err := snowflake.NewNode(machineID)
	if err != nil {
		return fmt.Errorf("snowflake: init node failed: %w", err)
	}
	node = n
	return nil
}

// GenID 生成一个全局唯一的 int64 ID
func GenID() int64 {
	return node.Generate().Int64()
}
