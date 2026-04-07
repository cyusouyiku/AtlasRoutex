package main

import "atlas-routex/internal/bootstrap"

func initializeApp() *bootstrap.App {
cfg := bootstrap.LoadConfig()
return bootstrap.NewApp(cfg)
}
