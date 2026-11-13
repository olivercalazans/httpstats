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
	"fmt"
	"os"
	"sync"

	"golang.org/x/term"
)



const (
	Flag4XX = "\033[31m4XX\033[0m" // red
	Flag3XX = "\033[33m3XX\033[0m" // orange
	Flag2XX = "\033[32m2XX\033[0m" // green
)



type Progress struct {
	mu      sync.Mutex
	jobs    int
	total   int
	done    int
	code2XX int
	code3XX int
	code4XX int
	errs    int
	isTTY   bool
}



func NewProgress(total, jobs int) *Progress {
	return &Progress{
		jobs:  jobs,
		total: total,
		isTTY: term.IsTerminal(int(os.Stdout.Fd())),
	}
}



func (p *Progress) Render() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rendLine()
}



func (p *Progress) rendLine() {
	if !p.isTTY { return }

	line := fmt.Sprintf("\033[K[...] Scanning with %d jobs %d/%d", p.jobs, p.done, p.total)
	line += fmt.Sprintf(" - [%s]: %d", Flag2XX, p.code2XX)
	line += fmt.Sprintf(" - [%s]: %d", Flag3XX, p.code3XX)
	line += fmt.Sprintf(" - [%s]: %d", Flag4XX, p.code4XX)
	line += fmt.Sprintf(" - [ERR]: %d", p.errs)
	
	fmt.Fprint(os.Stdout, "\r"+line)
}



func (p *Progress) Advance() {
	p.mu.Lock()
	p.done++
	p.rendLine()
	p.mu.Unlock()
}



func (p *Progress) Log(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isTTY {
		fmt.Fprint(os.Stdout, "\r\033[K")
	}

	fmt.Fprintln(os.Stdout, message)
	p.rendLine()
}



func (p *Progress) UpdateStats(code int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch {
	case code == 0:   p.errs++
	case code >= 400: p.code4XX++
	case code >= 300: p.code3XX++
	case code >= 200: p.code2XX++
	}
}



func (p *Progress) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isTTY {
		fmt.Fprintln(os.Stdout)
		return
	}

	fmt.Fprintf(os.Stdout,
		"done: %d/%d - [2XX]: %d - [3XX]: %d - [4XX]: %d - [ERR]: %d\n",
		p.done, p.total, p.code2XX, p.code3XX, p.code4XX, p.errs)
}