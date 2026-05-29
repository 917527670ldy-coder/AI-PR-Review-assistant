package server

func Start() error {
    r := NewRouter()
    return r.Run(":8081")
}
