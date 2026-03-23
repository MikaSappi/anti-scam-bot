package main

import (
    "encoding/json"
    "fmt"
    "io"
    "math/rand"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"
)

type Name struct {
    First []string `json:"first"`
    Last []string `json:"last"`
    Email []string `json:"email"`
}

type Person struct {
    FullName string
    Email string
}

var userAgents = []string{
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Safari/605.1.15",
    "Mozilla/5.0 (X11; Linux x86_64; rv:123.0) Gecko/20100101 Firefox/123.0",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:123.0) Gecko/20100101 Firefox/123.0",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
}

func randomUA() string {
    return userAgents[rand.Intn(len(userAgents))]
}

// submitForm POSTs fields to targetURL as application/x-www-form-urlencoded
// Returns (responseBody, statusCode, error).
func submitForm(targetURL string, fields map[string]string) (string, int, error) {
    formData := url.Values{}
    for k, v := range fields {
        formData.Set(k, v)
    }

    client := &http.Client{
        Timeout: 15 * time.Second,
    }

    req, err := http.NewRequest(http.MethodPost, targetURL, strings.NewReader(formData.Encode()))
    if err != nil {
        return "", 0, fmt.Errorf("building request: %w", err)
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("User-Agent", randomUA())

    resp, err := client.Do(req)
    if err != nil {
        return "", 0, fmt.Errorf("sending request: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", resp.StatusCode, fmt.Errorf("reading response: %w", err)
    }

    return string(body), resp.StatusCode, nil
}

func main() {

    // set up and read json
    jsonFile, err := os.Open("names.json")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
    fmt.Printf("Opened file.\n")

    defer jsonFile.Close()

    data, err := io.ReadAll(jsonFile)
    if err != nil {
        fmt.Printf("Couldn't read, because of %v\n", err)
    }
    fmt.Printf("read file. (in the past tense)\n")

    var names Name
    err = json.Unmarshal(data, &names)
    if err != nil {
        fmt.Printf("no: %v\n", err)
    }

    // target url and user agent
    targetURL := "https://example.com"
    fieldMap := func(p Person) map[string] string {
        return map[string]string{
            "name": p.FullName,
            "email": p.Email,
            "message": "Hi, could you please contact me as soon as possible?",
        }
    }

    // range through names
    people := make([]Person, 10)
    for i := range people {
        name := buildFullName(names)
        people[i] = Person{FullName: name, Email: buildEmail(name, names)}
    }

    for _, p := range people {
       fmt.Printf("Submitting as: %s <%s>\n", p.FullName, p.Email)

       _, status, err := submitForm(targetURL, fieldMap(p))
       if err != nil {
           fmt.Printf("  Error: %v\n", err)
       } else {
           fmt.Printf("  Status: %d\n", status)
       }

       isNight := time.Now().Hour() > 19 || time.Now().Hour() < 7

       var sleepFor time.Duration
       if isNight {
           sleepFor = time.Duration(rand.Intn(600)) * time.Second
       } else {
           sleepFor = time.Duration(rand.Intn(60)) * time.Second
       }

       fmt.Printf("  Sleeping %s\n", sleepFor)
       time.Sleep(sleepFor)
   }

}

func buildFullName(names Name) string {

    randFirst := names.First[rand.Intn(len(names.First))]
    randLast := names.Last[rand.Intn(len(names.Last))]

    return randFirst + " " + randLast
}

func buildEmail(n string, names Name) string {
    emailString := strings.ReplaceAll(n, " ", ".")
    emailString = strings.ToLower(emailString)

    randPredicate := rand.Intn(9)
    randInt := rand.Intn(90) + 10

    randEmailDomain := names.Email[rand.Intn(len(names.Email))]

    if randPredicate <= 5 {
        emailString += strconv.Itoa(randInt)
    }

    emailString = emailString + "@" + randEmailDomain

    return emailString
}
