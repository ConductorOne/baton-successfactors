package main

import (
	"github.com/conductorone/baton-sdk/pkg/config"
	cfg "github.com/conductorone/baton-successfactors/pkg/config"
)

func main() {
	config.Generate("successfactors", cfg.Config)
}
