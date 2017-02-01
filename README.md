# elblog [![GoDoc](https://godoc.org/github.com/piotrkowalczuk/elblog?status.svg)](https://godoc.org/github.com/piotrkowalczuk/elblog)

[![Build Status](https://travis-ci.org/piotrkowalczuk/elblog.svg?branch=master)](https://travis-ci.org/piotrkowalczuk/elblog)&nbsp;[![codecov](https://codecov.io/gh/piotrkowalczuk/elblog/branch/master/graph/badge.svg)](https://codecov.io/gh/piotrkowalczuk/elblog)


Library helps to parse [ELB](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/access-log-collection.html) logs. 

Elastic Load Balancing provides access logs that capture detailed information about requests sent to your load balancer. 
Each log contains information such as the time the request was received, the client's IP address, latencies, request paths, and server responses. 
You can use these access logs to analyze traffic patterns and to troubleshoot issues.

## Example 


```golang
package main 

import (
	"os"
	"fmt"
	
	"github.com/piotrkowalczuk/elblog"
)

func main() {
    file, err := os.Open("data.log")
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
    dec := elblog.NewDecoder(file)
    
    if dec.More() {
        log, err := dec.Decode()
        if err != nil {
            fmt.Println(err.Error())
            os.Exit(1)
        }
        fmt.Println(log)
    }
}
```