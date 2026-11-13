<h1 align="center">HTTPStats</h1>

A concurrent HTTP status scanner written in Go. Given a list of domains (or a single URL), it performs HTTP requests in parallel and reports each response's status code, highlighting 2XX/3XX/4XX classes and flagging HTML responses. It also supports searching and highlighting arbitrary terms in the output.

<br>

## Installation

### 1. Pre-built Linux binary

A pre-built static binary for Linux x86-64 is available on the Releases page (https://github.com/olivercalazans/httpstats/releases). No runtime dependencies, no shared libraries, no glibc version to worry about download.

```bash
curl -L -o httpstats https://github.com/olivercalazans/httpstats/releases/download/latest/httpstats
chmod +x httpstats
```

### 2. Other platforms or manual build
For any other platform (macOS, Windows, Linux ARM, etc.), or if you prefer to build from source, clone the project and compile it with Go:

```bash
git clone https://github.com/olivercalazans/httpstats.git
cd httpstats
go build -o httpstats
```

To produce a fully static, stripped binary -- no dynamic linker, no libc dependency, portable across any Linux distribution including Alpine and scratch containers -- use:

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o httpstats
```

Verify the result is statically linked:

```bash
file httpstats    # -> "ELF 64-bit ... statically linked"
ldd httpstats     # -> "not a dynamic executable"
```

<br>

## What it does

- Parallel scanning with bounded concurrency. Takes a list of domains (-f) or a single URL (-u) and dispatches HTTP requests across a configurable pool of workers (-j). Every goroutine is accounted for, and Ctrl+C triggers a clean shutdown -- in-flight requests are cancelled, the dispatcher stops, and no goroutine leaks.

- Per-request status classification. Each response is bucketed as 2XX, 3XX, 4XX, or ERR, and tallied live in the status line.

- Live progress line. In an interactive terminal, a single line at the bottom updates in place with the running counters and the number of completed jobs. It never scrolls or floods the screen.

- TTY-aware output. When stdout is redirected to a file or a pipe, the animated progress line is suppressed entirely and only a clean, machine-readable summary is written at the end -- no ANSI escapes or carriage returns polluting your logs.

- Path probing. With -p, every domain is tested against a specific path (/.git, /robots.txt, /admin, etc.) instead of the root.

- Redirect control. Redirects are not followed by default, so a 301 is reported as a 301 -- useful for mapping the true shape of a target. Pass -r to follow them.

- HTML detection. Responses whose Content-Type is text/html are tagged with an HTML badge, so you can spot web apps at a glance.

- Search-term highlighting. Pass -s "term1,term2" and every occurrence in the output is wrapped in yellow ANSI -- handy when scanning for exposed admin panels, git artifacts, or error strings.

- Literal search, not regex. Search terms are escaped before matching, so -s "a.b" matches the literal a.b rather than "a, any character, b". The tool never interprets your input as a pattern.

- Verbose mode. By default, only 2XX responses are printed. With -v, 3XX, 4XX, and network errors (timeouts, connection failures) are shown too.

- Configurable timeout. Every request is capped at a fixed timeout, so a single dead host can never stall the scan.

<br>

## Example

```bash
# Single URL, verbose
./httpstats -u example.com -v

# Batch from file, 10 concurrent workers, allow redirects
./httpstats -f domains.txt -j 10 -r

# Probe a specific path on every domain, highlight matches
./httpstats -f domains.txt -p /.git -s "git,php,md"
```

<br>

## Disclaimer

> [!WARNING]
> Use this software at your own risk!
>
> YOU AGREE TO:
> 
>   1. Use only with proper authorization
>   2. Comply with all applicable laws
>   3. Assume full liability for misuse
>
> The Developer assume NO liability.

<br>

## License
This project is licensed under the Apache-2.0 License. See the [LICENSE](./LICENSE) file for details.
