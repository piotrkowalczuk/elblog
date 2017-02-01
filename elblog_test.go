package elblog_test

import (
	"bufio"
	"fmt"
	"os"

	"reflect"
	"testing"

	"net"
	"net/http"
	"time"

	"bytes"
	"runtime"

	"github.com/piotrkowalczuk/elblog"
)

func Example() {
	file, err := os.Open("data.log")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	if scanner.Scan() {
		log, err := elblog.Parse(scanner.Bytes())
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println(log)
	}

	// Output:
	// &{2015-05-13 23:39:43.945958 +0000 UTC my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 73µs 1.048ms 57µs 200 200 0 29 GET http://www.example.com:80/ HTTP/1.1 curl/7.38.0 - -}
}

func TestParse(t *testing.T) {
	cases := map[string]struct {
		given    string
		expected elblog.Log
	}{
		"basic": {
			given: `2015-05-13T23:39:43.945958Z my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 0.000073 0.001048 0.000057 200 200 0 29 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.38.0" - -`,
			expected: elblog.Log{
				Time: func() time.Time {
					t, _ := time.Parse(time.RFC3339, "2015-05-13T23:39:43.945958Z")
					return t
				}(),
				Name: "my-loadbalancer",
				From: &net.TCPAddr{
					IP:   net.ParseIP("192.168.131.39"),
					Port: 2817,
				},
				To: &net.TCPAddr{
					IP:   net.ParseIP("10.0.0.1"),
					Port: 80,
				},
				RequestProcessingTime: func() time.Duration {
					d, _ := time.ParseDuration("73µs")
					return d
				}(),
				BackendProcessingTime: func() time.Duration {
					d, _ := time.ParseDuration("1.048ms")
					return d
				}(),
				ResponseProcessingTime: func() time.Duration {
					d, _ := time.ParseDuration("57µs")
					return d
				}(),
				ELBStatusCode:     http.StatusOK,
				BackendStatusCode: http.StatusOK,
				ReceivedBytes:     0,
				SentBytes:         29,
				Request:           "GET http://www.example.com:80/ HTTP/1.1",
				UserAgent:         "curl/7.38.0",
				SSLCipher:         "-",
				SSLProtocol:       "-",
			},
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			got, err := elblog.Parse([]byte(c.given))
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if !reflect.DeepEqual(*got, c.expected) {
				t.Errorf("expected:\n	%v but got:\n	%v", c.expected, *got)
			}
		})
	}
}

var benchLog elblog.Log

func BenchmarkParse(b *testing.B) {
	data := []byte(`2015-05-13T23:39:43.945958Z my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 0.000073 0.001048 0.000057 200 200 0 29 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.38.0" - -`)
	for n := 0; n < b.N; n++ {
		log, err := elblog.Parse(data)
		if err != nil {
			b.Fatalf("unexpected error: %s", err.Error())
		}
		benchLog = *log
	}
}

func buffor(max int) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	for i := 0; i < max; i++ {
		buf.WriteString(`2015-05-13T23:39:43.945958Z my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 0.000073 0.001048 0.000057 200 200 0 29 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.38.0" - -`)
		buf.WriteRune('\n')
	}
	return buf
}

func BenchmarkParse_NonParallel(b *testing.B) {
	buf := buffor(100000)
	b.ResetTimer()

	for n := 0; n <= b.N; n++ {
		b.StopTimer()
		buff := *buf
		scanner := bufio.NewScanner(&buff)
		scanner.Split(bufio.ScanLines)
		b.StartTimer()

		if scanner.Scan() {
			log, err := elblog.Parse(scanner.Bytes())
			if err != nil {
				b.Fatalf("unexpected error: %s", err.Error())
			}

			benchLog = *log
		}
	}
}

func BenchmarkParse_Parallel(b *testing.B) {
	buf := buffor(100000)
	parallelism := runtime.NumCPU()*10
	b.ResetTimer()

	for n := 0; n <= b.N; n++ {
		b.StopTimer()

		buff := *buf
		scanner := bufio.NewScanner(&buff)
		scanner.Split(bufio.ScanLines)

		in := make(chan []byte)
		out := make(chan *elblog.Log)
		done := make(chan error, parallelism+1)

		for i := 0; i < parallelism; i++ {
			go func(in <-chan []byte, out chan<- *elblog.Log, done chan<- error) {
				for b := range in {
					log, err := elblog.Parse(b)
					if err != nil {
						done <- err
					}
					out <- log
				}
				done <- nil
			}(in, out, done)
		}

		go func(out <-chan *elblog.Log, done chan<- error) {
			for log := range out {
				benchLog = *log
			}
			done <- nil
		}(out, done)

		b.StartTimer()

		if scanner.Scan() {
			in <- scanner.Bytes()
		}
		close(in)

		kill := 0
	DoneLoop:
		for err := range done {
			if err != nil {
				b.Fatalf("unexpected error: %s", err.Error())
			}
			kill++
			if kill == parallelism {
				close(out)
				break DoneLoop
			}
		}
	}
}
