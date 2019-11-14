package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Create a new type for a list of Strings
type stringList []string

// Implement the flag.Value interface
func (s *stringList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringList) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

func main() {
	// res := make(map[string]int)

	text := flag.String("text", "", "Text to find. (Required)")
	var urlsArray stringList
	flag.Var(&urlsArray, "urls", "List of urls to parse. (Required)")

	flag.Parse()

	if *text == "" || (&urlsArray).String() == "[]" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("text: %s, urls: %v\n",
		*text,
		&urlsArray,
	)

	inputCh := createInputCh(urlsArray)

	rout1 := getRequest(inputCh, *text)
	rout2 := getRequest(inputCh, *text)
	rout := fanIn(rout1, rout2)

	for i := 0; i < len(urlsArray); i++ {
		fmt.Println("res", <-rout)
	}

	// fmt.Println(res)
	// for k, v := range res {
	// 	fmt.Printf("%s - %d\n", k, v)
	// }
}

func createInputCh(urls []string) <-chan string {
	out := make(chan string)
	go func() {
		for _, u := range urls {
			out <- u
		}
		close(out)
	}()
	return out
}

func getRequest(in <-chan string, word string) <-chan int {
	fmt.Println("start")
	out := make(chan int)
	go func() {
		for n := range in {
			// fmt.Println(n)
			var client http.Client
			url := "http://" + n
			resp, err := client.Get(url)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					os.Exit(1)
				}
				bodyString := string(bodyBytes)
				res := 0
				reg, err := regexp.Compile("[^a-zA-Z0-9]+")
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
				processedString := reg.ReplaceAllString(bodyString, " ")
				arr := strings.Split(processedString, " ")
				for i := 0; i < len(arr); i++ {
					if arr[i] == word {
						res += 1
					}
				}
				fmt.Println("res", n, res)
				out <- res
			}
		}
		close(out)
	}()
	time.Sleep(2 * time.Second)
	fmt.Println("end")
	return out
}

func fanIn(input1, input2 <-chan int) <-chan int {
	c := make(chan int)
	go func() {
		for {
			select {
			case s := <-input1:
				c <- s
			case s := <-input2:
				c <- s
			}
		}
	}()
	return c
}

// "go.lintTool":"golangci-lint",
// "go.lintFlags": [
//   "--fast"
// ]
