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

package internal

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
)



type Args struct {
	File    string
	URL     string
	Path    string
	Search  string

	Redirection bool
	Verbose     bool
	Jobs 		int
}



func NewArgs() *Args {
	args := &Args{}

	args.parseCLIArgs()
	args.validJobs()
	args.validSearch()

	return args
}



func (a *Args) parseCLIArgs() {
	flag.StringVar(&a.File, "f", "", "TXT file with domain list")
	flag.StringVar(&a.URL, "u", "", "Check only one URL")
	flag.StringVar(&a.Path, "p", "", "URL path to check on each domain (e.g., \"/.git\" or \"/robots.txt\")")
	flag.StringVar(&a.Search, "s", "", "One or more terms to search and highlight in yellow (e.g., term1,term2)")

	flag.BoolVar(&a.Redirection, "r", false, "Allow redirection")
	flag.BoolVar(&a.Verbose, "v", false, "DEFAULT: Only 200")
	
	flag.IntVar(&a.Jobs, "j", 5, "Number of parallel jobs to run (DEFAULT: 5)")

	flag.Parse()
}



func (a *Args) validJobs() {
	if a.Jobs <= 0 {
		Fatal(fmt.Sprintf("Invalid number for jobs (%d). It must be 1 or higher", a.Jobs))
	}
}



func (a *Args) validSearch() {
	if a.Search == "" { return }

	terms   := strings.Split(a.Search, ",")
	escaped := make([]string, len(terms))
	
	for i, term := range terms {
		escaped[i] = regexp.QuoteMeta(term)
	}

	a.Search = "(" + strings.Join(escaped, "|") + ")"
}