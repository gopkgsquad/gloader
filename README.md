# gloader

<p align="center">
  gloader is a package built for live reload. Designed for easy use.
</p>

## ⚙️ Installation

```bash
go get -u github.com/gopkgsquad/gloader
```

## Quickstart

```go
package main

import "github.com/gopkgsquad/gloader"

func main() {
    // initialize a new http ServeMux
    router := http.NewServeMux()

    // initialize http.Server
	srv := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}

    // start the application with live reload
    gloader.NewWatcher(srv, time.Second*2).Start()

}
```

## Examples

```go
func main() {
    // initialize a new glogger
    logger := logger.NewLogger(os.Stdout, logger.LogLevelInfo)

    // log info
    logger.Info("FROM MYLOGGER INFO")

    // log Warning
    logger.Warning("FROM MYLOGGER WARNING")

    // log Error
    logger.Error("FROM MYLOGGER ERROR")

    // log Fatal
    logger.Fatal("FROM MYLOGGER FATAL")
}

```
