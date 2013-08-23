package amzpa

import (
    "log"
    "testing"
    "net/http"
    "os"
)

func TestSearch(t *testing.T) {
    req := NewRequest(
        os.Getenv("AWSKey"),
        os.Getenv("AWSSecret"),
        os.Getenv("ATag"),
        "US",
        &http.Client{},
    )
    ret, _ := req.Search("Romance", "Books", "Images,Small", "salesrank")
    log.Println(string(ret))
}

func TestLookup(t *testing.T){
    req := NewRequest(
        os.Getenv("AWSKey"),
        os.Getenv("AWSSecret"),
        os.Getenv("ATag"),
        "US",
        &http.Client{},
    )

    ret, _ := req.Lookup([]string{"B007HCCNJU"}, "Accessories", "ASIN")
    log.Println(string(ret))
}
