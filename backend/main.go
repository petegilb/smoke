package main

import (
    "fmt"
    "net/http"
    "os"
    "io"
)

func main() {
    fmt.Println("Hello, World!")

    requestURL := fmt.Sprintf("https://www.valvesoftware.com/about/stats")
    res, err := http.Get(requestURL)
    if err != nil {
        fmt.Printf("error making http request: %s\n", err)
        os.Exit(1)
    }

    resBody, err := io.ReadAll(res.Body)
    if err != nil {
        fmt.Printf("client: could not read response body: %s\n", err)
        os.Exit(1)
    }
    fmt.Printf("client: response body: %s\n", resBody)

    fmt.Printf("client: got response!\n")
    fmt.Printf("client: status code: %d\n", res.StatusCode)
}
