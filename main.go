package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
	// "context"
)

type stringList []string

type Result struct {
	url   string
	count int
}

func (s *stringList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringList) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

func main() {
	// var t *testing.T
	// TestProg(t)

	res := make(map[string]int)

	text := flag.String("text", "", "Text to find. (Required)")
	var urlsArray stringList
	flag.Var(&urlsArray, "urls", "List of urls to parse. (Required)")

	flag.Parse()

	if *text == "" || (&urlsArray).String() == "[]" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// fmt.Printf("text: %s, urls: %v\n",
	// 	*text,
	// 	&urlsArray,
	// )

	killsignal := make(chan bool)
	queue := make(chan Result)
	done := make(chan bool)

	numberOfWorkers := 2
	for i := 0; i < numberOfWorkers; i++ {
		go worker(queue, i, done, killsignal, res)
	}

	for _, u := range urlsArray {
		go getRequest(queue, u, text)
	}

	for c := 0; c < len(urlsArray); c++ {
		<-done
	}

	close(killsignal)
	time.Sleep(2 * time.Second)
	for val := range res {
		if res[val] == -1 {
			fmt.Printf("%s - wrong url\n", val)
			continue
		}
		fmt.Printf("%s - %d\n", val, res[val])
	}
}

func getRequest(q chan Result, u string, word *string) {
	var client http.Client
	url := "http://" + u
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		q <- Result{u, -1}
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		os.Exit(1)
	}
	bodyString := string(bodyBytes)
	res := Result{u, 0}
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	processedString := reg.ReplaceAllString(bodyString, " ")
	arr := strings.Split(processedString, " ")
	for i := 0; i < len(arr); i++ {
		if arr[i] == *word {
			res.count += 1
		}
	}
	q <- res
}

func worker(queue chan Result, worknumber int, done, ks chan bool, res map[string]int) {
	for true {
		select {
		case k := <-queue:
			res[k.url] = k.count
			done <- true
		case <-ks:
			return
		}
	}
}

func TestProg(t *testing.T) {
	type testCase struct {
		url   string
		word  string
		Count int
	}
	cases := []testCase{
		{"ru.wikipedia.org/wiki/Go", "go", 83},
		{"medium.com/rungo/working-in-go-workspace-3b0576e0534a", "go", 132},
		{"azaza", "azaza", -1},
	}

	for _, test := range cases {
		var res Result
		log.Print("Start GET ", test.url, test.word)
		inpCh := make(chan Result)
		go getRequest(inpCh, test.url, &test.word)
		res = <-inpCh

		if test.Count != res.count {
			log.Print("Fail\n")
			t.Fail()
			continue
		}

		log.Print("Success\n")
	}
}
