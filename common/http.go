package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	jsonq "github.com/jmoiron/jsonq"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

// ******************
// * POST FUNCTIONS *
// ******************

// Perform an HTTP PUT on the supplied url with the body of the supplied object reference
func PostJson(url string, body interface{}) string {
	b, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Can't marshall body attempting to call %si\n", url)
		panic(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Printf("Error trying to call URL: %s\n", url)
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body from: %s\n", url)
		panic(err)
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		return string(responseBody)
	} else if resp.StatusCode == 409 {
		fmt.Println("Got a 409 response, maybe exploud is being deployed?")
		fmt.Println(string(responseBody))
	}
	fmt.Println(string(responseBody))
	panic(fmt.Sprintf("%d response calling URL: ", resp.StatusCode))
}

// Posts the supplied JSON to the url and unmarshals the response to the supplied
// struct.
func PostJsonUnmarshalResponse(url string, body interface{}, v interface{}) {
	responseBody := PostJson(url, &body)
	err := json.Unmarshal([]byte(responseBody), &v)
	if err != nil {
		fmt.Printf("Unable to unmarshal response from %s\n", url)
		panic(err)
	}
}

// *****************
// * PUT FUNCTIONS *
// *****************

func PutByteArray(url string, data []byte) string {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		fmt.Printf("Error making PUT request to url: %s\n", url)
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Error trying to call URL: %s\n", url)
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body from: %s\n", url)
		panic(err)
	}

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		return string(responseBody)
	}
	panic(fmt.Sprintf("Got %d response calling: %s with body: %s.\nResponse was: %s",
		resp.StatusCode, url, data, string(responseBody)))
}

// Performs an HTTP PUT on the supplied url with the body of the supplied string.
func PutString(url string, body string) string {
	return PutByteArray(url, []byte(body))
}

// Performs an HTTP PUT on the supplied url with the body of the supplied object reference
func PutJson(url string, body interface{}) string {
	b, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Unable to Marshall json for http put to url %s\n", url)
		panic(err)
	}
	return PutByteArray(url, b)
}

// *****************
// * GET FUNCTIONS *
// *****************

// Performs an HTTP GET request on the supplied url and returns the result
// as a string. Returns non nil err on non 200 response. Optionally takes
// var args of func(*http.Request) that can be used to mute the headers
func GetString(url string, muteRequest ...func(*http.Request)) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating GET request for url: %s\n", url)
		panic(err)
	}
	for i := range muteRequest {
		muteRequest[i](req)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Error trying to call URL: %s\n", url)
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body from: %s\n", url)
		panic(err)
	}

	if resp.StatusCode == 200 {
		return string(responseBody)
	}
	panic(fmt.Sprintf("Got %d response calling: %s \nResponse was: %s",
		resp.StatusCode, url, string(responseBody)))
}

// Returns the result of an http get request as a jsonq object
func GetAsJsonq(url string) *jsonq.JsonQuery {
	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(GetString(url)))
	dec.Decode(&data)
	return jsonq.NewQuery(data)
}

// Performs an HTTP GET request on the supplied url and unmarshals the response
// into the supplied object. Returns non nil error on failure
func GetJson(url string, v interface{}) {
	err := json.Unmarshal([]byte(GetString(url)), &v)
	if err != nil {
		fmt.Printf("Unable to marshall response from url %s\n", url)
		panic(err)
	}
}

// ******************
// * HEAD FUNCTIONS *
// ******************

// Performs an HTTP HEAD call on the supplied URL. Returns true if
// the response code is 200.
func Head(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Printf("Error calling head on URL: %s\n", url)
		panic(err)
	}
	switch resp.StatusCode {
	case 200:
		return true
	case 404:
		return false
	default:
		panic(fmt.Sprintf("Unknown response: %d from HEAD on URL: %s",
			resp.StatusCode, url))
	}
}

// *****************
// * CUSTOM CLIENT *
// *****************

// Creates a new http.Client with a longer timeout, optionally accepts a Config object
func NewTimeoutClient(connect time.Duration, read time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{

			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, connect)
				if err != nil {
					return nil, err
				}

				conn.SetDeadline(time.Now().Add(read))
				conn.SetReadDeadline(time.Now().Add(read))
				conn.SetWriteDeadline(time.Now().Add(read))
				return conn, nil
			},

			Proxy: http.ProxyFromEnvironment,
		},
	}
}
