package main

import (
	"fmt"

	"github.com/elliot40404/uc/internal/config"
)

func (e env) init() error {
	if err := config.Write(e.ConfigPath, config.Starter(e.Found(), e.Home), e.Home); err != nil {
		return err
	}
	fmt.Println("wrote", e.ConfigPath)
	return nil
}
