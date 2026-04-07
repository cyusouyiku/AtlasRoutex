package main

import (
"log"

"atlas-routex/internal/bootstrap"
)

func main() {
cfg := bootstrap.LoadConfig()
app := bootstrap.NewApp(cfg)
srv := bootstrap.NewHTTPServer(cfg.HTTPAddr, app.BuildMux())
log.Printf("api listening on %s", cfg.HTTPAddr)
if err := srv.Start(); err != nil {
log.Fatal(err)
}
}
