<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![Tests][tests-shield]][tests-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">

<h3 align="center">iqhater/pkg</h3>

  <p align="center">
    A lightweight collection of reusable Go packages for HTTP services, middleware, background processing, and simple utilities.
    <br />
    <br />
    <a href="https://github.com/iqhater/pkg"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/iqhater/pkg">View Project</a>
    ·
    <a href="https://github.com/iqhater/pkg/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    ·
    <a href="https://github.com/iqhater/pkg/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#examples">Examples</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->
## About The Project

`iqhater/pkg` is a small, framework-free Go library with reusable primitives for API services.

It focuses on common infrastructure pieces without adding routers, dependency injection, database layers, or application scaffolding. Use only the packages you need.

### Built With

* [![Go][go.dev]][Go-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

Install Go and, optionally, Taskfile for local development commands.

* [Go](https://go.dev/dl/)
* [Taskfile](https://taskfile.dev/installation/) _(recommended)_
* [golangci-lint](https://golangci-lint.run/welcome/install/) _(optional)_

### Installation

1. Clone the repo

   ```sh
   git clone https://github.com/iqhater/pkg.git
   ```

2. Install Go module dependencies and development tools

   ```sh
   task install
   ```

3. Run tests

   ```sh
   task test
   ```

4. Run the example server

   ```sh
   task run
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

Install the package in your project:

```sh
go get github.com/iqhater/pkg
```

Combine middleware with the standard `net/http` package:

```go
package main

import (
	"net/http"
	"time"

	"github.com/iqhater/pkg/headers"
	"github.com/iqhater/pkg/middleware"
)

func main() {
	cache := middleware.NewCache("10s")

	mid := middleware.Middlewares(
		middleware.Recover,
		middleware.RequestID,
		headers.CORSHeaders(headers.CORSConfig{
			AllowOrigins: []string{"http://localhost:4000"},
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
			AllowHeaders: []string{"Accept", "Content-Type", "Authorization"},
		}),
		middleware.Log,
		middleware.Limit(2, 5),
		middleware.ContextTimeout(3*time.Second),
		headers.SecureHeaders,
		middleware.Compress,
		cache.CacheResponse,
		headers.ContentTypeHeaders("application/json"),
	)

	http.HandleFunc("GET /api", middleware.Bind(mid, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	}))

	http.ListenAndServe(":4000", nil)
}
```

### Taskfile useful commands

Run all tests

```sh
task test
```

Run the project linter

```sh
task lint
```

Build the example binary

```sh
task build
```

Remove example binaries

```sh
task clean
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- EXAMPLES -->
## Examples

Runnable examples are available in the `examples` package.

```sh
go run .\examples\middlewares\main.go
```

Example package overview:

* `middleware` — recovery, request IDs, logging, rate limiting, timeouts, caching, and compression.
* `headers` — CORS, security headers, and content type headers.
* `async/workerpool` — generic worker pool for concurrent jobs.
* `async/eventbus` — simple in-process event bus.
* `generate` — random token generation.

### Notes

* `CacheResponse` caches responses by request URI and only stores successful `200 OK` responses.
* `Compress` prefers Brotli when the client sends `Accept-Encoding: br`, otherwise it falls back to gzip.
* `Compress` skips WebSocket upgrade requests.
* `ContextTimeout` cancels the request context, but handlers must explicitly check `r.Context().Done()` for long-running work.
* `RequestID` stores the request ID in request context. Read it from the request passed to the downstream handler.
* `Limit` is based on `req.RemoteAddr`, so it is useful for simple per-IP rate limiting.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag `enhancement`.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/amazing-feature`)
3. Commit your Changes (`git commit -m 'Add amazing feature'`)
4. Push to the Branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->
## Contact

email - <iqhater@yandex.ru>

Project Link: [https://github.com/iqhater/pkg](https://github.com/iqhater/pkg)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[issues-shield]: https://img.shields.io/github/issues/iqhater/pkg.svg?style=for-the-badge
[issues-url]: https://github.com/iqhater/pkg/issues
[license-shield]: https://img.shields.io/github/license/iqhater/pkg.svg?style=for-the-badge
[license-url]: https://github.com/iqhater/pkg/blob/main/LICENSE.txt
[tests-url]: https://github.com/iqhater/pkg/actions/workflows/run_ci_tests.yml/badge.svg
[tests-shield]: https://github.com/iqhater/pkg/actions/workflows/run_ci_tests.yml/badge.svg
[go.dev]: https://img.shields.io/badge/golang-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://go.dev
