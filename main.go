// Copyright 2026 Oliver R. Calazans Jeronimo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"context"
	"fmt"
	"httpstats/internal"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)



const requestTimeout = 3 * time.Second



type HTTPStats struct {
    args      *internal.Args
    urlList   []string
    progress  *internal.Progress
    re        *regexp.Regexp
}



func NewScanner(args *internal.Args) *HTTPStats {
    s := &HTTPStats{args: args}

    if args.Search != "" {
        re, err := regexp.Compile("(?i)" + args.Search)
        if err != nil {
            internal.Fatal(fmt.Sprintf("Invalid search pattern: %v", err))
        }
        s.re = re
    }

    return s
}



func (s *HTTPStats) Run() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := s.loadURLs(); err != nil {
        internal.Fatal(err.Error())
    }

    completed := s.scan(ctx)

    if !completed {
        fmt.Println("\nProcess stopped by the user")
    }
}



func (s *HTTPStats) loadURLs() error {
    if s.args.File != "" && s.args.URL != "" {
        return fmt.Errorf("use either -f or -u, not both")
    }

    if s.args.URL != "" {
        s.urlList = []string{s.args.URL}
        return nil
    }

    if s.args.File == "" {
        return fmt.Errorf("either -f or -u is required")
    }

    f, err := os.Open(s.args.File)
    if err != nil {
        return fmt.Errorf("could not open file: %w", err)
    }
    defer f.Close()

    sc := bufio.NewScanner(f)
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        s.urlList = append(s.urlList, line)
    }

    if err := sc.Err(); err != nil {
        return fmt.Errorf("error reading file: %w", err)
    }

    if len(s.urlList) == 0 {
        return fmt.Errorf("URL list is empty")
    }

    return nil
}



func (s *HTTPStats) scan(ctx context.Context) bool {
    s.progress = internal.NewProgress(len(s.urlList), s.args.Jobs)
    s.progress.Render()

    client := &http.Client{Timeout: requestTimeout}
    if !s.args.Redirection {
        client.CheckRedirect = func(*http.Request, []*http.Request) error {
            return http.ErrUseLastResponse
        }
    }

    sem := make(chan struct{}, s.args.Jobs)
    var wg sync.WaitGroup

    completed := true

outer:
	for _, u := range s.urlList {
        select {
        case <-ctx.Done():
            completed = false
            break outer
        case sem <- struct{}{}:
        }

        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            defer func() { <-sem }()
            s.checkURL(ctx, client, u)
        }(u)
    }

    wg.Wait()
    s.progress.Finish()

    return completed
}



func (s *HTTPStats) checkURL(ctx context.Context, client *http.Client, raw string) {
    defer s.progress.Advance()

    url := s.formatURL(raw)

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        s.displayWarning(fmt.Sprintf("Invalid URL: %s", url))
        return
    }

    resp, err := client.Do(req)
    if err != nil {
        switch {
        case os.IsTimeout(err):
            s.displayWarning(fmt.Sprintf("Timeout: %s", url))
        default:
            s.displayWarning(fmt.Sprintf("Connection err: %v", err))
        }
        return
    }
    defer resp.Body.Close()

    s.displayResponse(resp)
}



func (s *HTTPStats) formatURL(url string) string {
    url = strings.TrimRight(url, "/")

    if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
        url = "https://" + url
    }

    if s.args.Path == "" {
        return url
    }

    url   = strings.TrimSuffix(url, s.args.Path)
    path := strings.Trim(s.args.Path, "/")

    return url + "/" + path
}



func (s *HTTPStats) displayResponse(resp *http.Response) {
    code := resp.StatusCode
    s.progress.UpdateStats(code)

    if !s.args.Verbose && code >= 300 {
        return
    }

    var x string
    switch {
    case code >= 400: x = fmt.Sprintf("\033[31m%d\033[0m", code)
    case code >= 300: x = fmt.Sprintf("\033[33m%d\033[0m", code)
    case code >= 200: x = fmt.Sprintf("\033[32m%d\033[0m", code)
    default:          x = fmt.Sprintf("%d", code)
    }

    var z string
    if internal.IsHTML(resp) {
        z = " \033[34mHTML\033[0m "
    } else {
        z = " "
    }

    final := resp.Request.URL.String()
    line  := fmt.Sprintf("[%s]%s%s", x, z, final)

    if s.re != nil {
        line = s.re.ReplaceAllString(line, "\033[33m$0\033[0m")
    }

    s.progress.Log(line)
}



func (s *HTTPStats) displayWarning(text string) {
    s.progress.UpdateStats(0)

    if s.args.Verbose {
        s.progress.Log(fmt.Sprintf("[!!!] %s", text))
    }
}



func main() {
	args 	  := internal.NewArgs()
	httpstats := NewScanner(args)

	httpstats.Run()
}