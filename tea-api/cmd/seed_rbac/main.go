package main

import (
	"flag"
	"fmt"

	"tea-api/internal/config"
	"tea-api/internal/service"
	"tea-api/pkg/database"
)

func main() {
	var cfgPath string
	var assignOpenID string
	flag.StringVar(&cfgPath, "config", "configs/config.yaml", "config file path")
	flag.StringVar(&assignOpenID, "assign-openid", "", "assign auditor role to this openid (optional)")
	flag.Parse()

	if err := config.LoadConfig(cfgPath); err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}
	database.InitDatabase()
	if err := service.SeedRBAC(database.GetDB(), service.SeedOptions{AssignOpenID: assignOpenID}); err != nil {
		panic(err)
	}
	fmt.Println("RBAC seed done")
}
