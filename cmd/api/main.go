package main

import (
    "log"

    "xengineer/internal/server"
)

func main() {
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
