package main

import (
	"fmt"

	"github.com/elliot40404/uc/internal/config"
)

func (e env) init() error {
	if err := config.Write(e.configPath, config.Starter(e.found(), e.home), e.home); err != nil {
		return err
	}
	fmt.Println("wrote", e.configPath)
	return nil
}
