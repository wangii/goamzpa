// Copyright 2012 Marco Dinacci. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package amzpa provides functionality for using the
// Amazon Product Advertising service.

package amzpa

import (
	"fmt"
	"sort"
	"time"
	"io/ioutil"
	"strings"
	// "net/url"
	"net/http"
	"encoding/base64"
	"crypto/hmac"
	"crypto/sha256"
    "bytes"
    // "log"
)

var b64 = base64.StdEncoding
var service_domains = map[string] string {
     "CA": "ecs.amazonaws.ca",
     "CN": "ecs.amazon.cn",
     "DE": "ecs.amazonaws.de",
     "ES": "ecs.amazon.es",
     "FR": "ecs.amazonaws.fr",
     "IT": "ecs.amazon.it",
     "JP": "ecs.amazonaws.jp",
     "UK": "ecs.amazonaws.co.uk",
     "US": "ecs.amazonaws.com",
}

type AmazonRequest struct {
	accessKeyID string
	accessKeySecret string
	associateTag string
	region string
    client *http.Client
}

// Create a new AmazonRequest initialized with the given parameters
func NewRequest(accessKeyID, accessKeySecret, associateTag, region string, client *http.Client) *AmazonRequest {
	return &AmazonRequest{accessKeyID, accessKeySecret, associateTag, region, client}
}

// Perform an ItemLookup request.
//
// Usage:
// ids := []string{"01289328","2837423"}
// response,err := request.ItemLookup(ids, "Medium,Accessories", "ASIN")
func (self *AmazonRequest) Lookup(itemIds []string, responseGroups string, idType string) ([]byte, error) {

	args := make(map[string]string)

	args["Operation"] = "ItemLookup"
	args["ItemId"] = strings.Join(itemIds, ",")
	args["ResponseGroup"] = responseGroups
	args["IdType"] = idType

	// Do request
	content, err := self.doRequest(self.buildRequest(args))

	if err != nil {
		return nil, err
	}

	return content, nil
}

func (self *AmazonRequest) Search(q, index, responseGroups, sort string, extra map[string]string) ([]byte, error) {

	args := make(map[string]string)

	args["Operation"] = "ItemSearch"
	args["SearchIndex"] = index

    if len(q) > 0 {
        args["Keywords"] = q
    }

	args["ResponseGroup"] = responseGroups
	args["Sort"] = sort
    if extra!= nil {
        for k,v := range extra {
            args[k] = v
        }
    }

	content, err := self.doRequest(self.buildRequest(args))

	if err != nil {
		return nil, err
	}

	return content, nil
}

func (self *AmazonRequest)buildRequest(args map[string]string) string {

	args["AWSAccessKeyId"] = self.accessKeyID
	args["Service"] = "AWSEcommerceService"
	args["AssociateTag"] = self.associateTag
	args["Version"] = "2011-08-01"
    // args["SignatureVersion"] = "2"
	// args["SignatureMethod"] = "HmacSHA256"

	args["Timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// Sort the keys otherwise Amazon hash will be
	// different from mine and the request will fail
	keys := make([]string, 0, len(args))

	for a := range args {
		keys = append(keys, a)
	}

	sort.Strings(keys)
    var qs bytes.Buffer

    kl := len(keys) - 1
	for idx, key := range keys {
		escaped := Encode(args[key])

		qs.WriteString(Encode(key))
        qs.WriteString("=")
        qs.WriteString(escaped)

        if idx < kl {
            qs.WriteString("&")
        }
	}

	// Hash & Sign
	domain := service_domains[self.region]

	payload := "GET\n" + domain + "\n/onca/xml\n" + qs.String()
	hash := hmac.New(sha256.New, []byte(self.accessKeySecret))
	hash.Write([]byte(payload))

    // sig := make([]byte, b64.EncodedLen(hash.Size()))
    // b64.Encode(sig, hash.Sum(nil))
    sig := b64.EncodeToString(hash.Sum(nil))

	qs.WriteString("&Signature=")
    qs.WriteString(Encode(sig))

    ret := fmt.Sprintf("http://%s/onca/xml?%s", domain, qs.String())
    return ret
}

// TODO add "Accept-Encoding": "gzip" and override UserAgent
// which is set to Go http package.
func (self *AmazonRequest)doRequest(requestURL string) ([]byte, error) {

    resp, err := self.client.Get(requestURL)

	if err != nil {
		return nil, err
	}

    contents, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	return contents, err
}

var unreserved = make([]bool, 128)
var hex = "0123456789ABCDEF"

func init() {
	// RFC3986
	u := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890-_.~"
	for _, c := range u {
		unreserved[c] = true
	}
}

// Encode takes a string and URI-encodes it in a way suitable
// to be used in AWS signatures.
func Encode(s string) string {
    encode := false
    for i := 0; i != len(s); i++ {
        c := s[i]
        if c > 127 || !unreserved[c] {
            encode = true
            break

        }

    }
    if !encode {
        return s

    }
    e := make([]byte, len(s)*3)
    ei := 0
    for i := 0; i != len(s); i++ {
        c := s[i]
        if c > 127 || !unreserved[c] {
            e[ei] = '%'
            e[ei+1] = hex[c>>4]
            e[ei+2] = hex[c&0xF]
            ei += 3

        } else {
            e[ei] = c
            ei += 1

        }
    }

    return string(e[:ei])
}
