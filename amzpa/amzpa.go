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
	"net/url"
	"net/http"
	"encoding/base64"
	"crypto/hmac"
	"crypto/sha256"
    "bytes"
    "log"
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
     // "CA": "ecs.amazonaws.ca",
     // "CN": "webservices.amazon.cn",
     // "DE": "ecs.amazonaws.de",
     // "ES": "webservices.amazon.es",
     // "FR": "ecs.amazonaws.fr",
     // "IT": "webservices.amazon.it",
     // "JP": "ecs.amazonaws.jp",
     // "UK": "webservices.amazon.co.uk",
     // "US": "ecs.amazonaws.com",
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
	content, err := self.doRequest(self.buildURL(args))

	if err != nil {
		return nil, err
	}

	return content, nil
}

func (self *AmazonRequest) Search(q, index, responseGroups, sort string) ([]byte, error) {

	args := make(map[string]string)

	args["Operation"] = "ItemSearch"
	args["SearchIndex"] = index
	args["Keywords"] = q
	args["ResponseGroup"] = responseGroups
	args["Sort"] = sort

	content, err := self.doRequest(self.buildURL(args))

	if err != nil {
		return nil, err
	}

	return content, nil
}

func (self *AmazonRequest)buildURL(args map[string]string) string {
    
    now := time.Now().UTC()
	args["AWSAccessKeyId"] = self.accessKeyID
	args["Service"] = "AWSEcommerceService"
	args["AssociateTag"] = self.associateTag
	args["Version"] = "2011-08-01"
    // args["Validate"] = "True"
    // args["SignatureVersion"] = "2"
	// args["SignatureMethod"] = "HmacSHA256"

	args["Timestamp"] = now.Format("2006-01-02T15:04:05Z")
    log.Println(args["Timestamp"])
    //time.RFC3339)

	// Sort the keys otherwise Amazon hash will be
	// different from mine and the request will fail
	keys := make([]string, 0, len(args))

	for a := range args {
		keys = append(keys, a)
	}

	sort.Strings(keys)

	// There's probably a more efficient way to concatenate strings, not a big deal though.
    var qs bytes.Buffer
    
    kl := len(keys) - 1
	for idx, key := range keys {
		escaped := url.QueryEscape(args[key])

		qs.WriteString(key)
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
    qs.WriteString(sig)

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
