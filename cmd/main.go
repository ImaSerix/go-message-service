package main

import (
	"fmt"
	"strconv"

	"github.com/ImaSerix/go-message-service/internal/generator"
)

func main() {
	gen := generator.NewSnowflakeIDGenerator(1, 0)
	fmt.Println(strconv.FormatInt(int64(gen.NewID()), 2))
}
