# Logging

The Bearer agent may emit warning or error information in some situations.

To make this logging simple to use in both simple and advanced usage scenarios,
the agent :

- accepts a `Logger` allowing the application to pass a logger of its choice;
- provides a default logger used if the application does not define one.

For compatibility with the runtime `log` package, the logger accepts any `io.Writer`
implementation. If the writer is a rs/zerolog `Logger`, it will be used as such, 
otherwise a new `Logger` instance will be created to wrap the plain `Writer`.


## Design: Logger signature choice

The Go standard `log` package does not handle structured logging, but allows
configuration of its output format, all output going to an `io.Writer` instance,
which defaults to `os.Stderr`.

As the most widely used logger in Go code, the logger used by the Bearer agent
needs to support it.


### Terminology

A _structured_ logger is able to emit log field, not just messages and timestamps.
A _leveled_ logger supports adding a severity level to all messages.


### uber/zap

Zap defines 7 logging levels, which do not really map to RFC5424:

| RFC5424 | Name                        | Zap  | Zap name
|:-------:|-----------------------------|:----:|----------------|
| 7       | Debug                       | -1   | Debug          |       
| 6       | Informational               | 0    | Informational  |
| 5       | Notice                      |      | Notice         |
| 4       | Warning                     | 1    | Warning        |
| 3       | Error                       | 2    | Error          |
|         |                             | 3    | DPanic         |
| 2       | Critical                    | 4    | Panic          |
| 1       | Alert                       | 5    | Fatal          |
| 0       | Emergency: system unusable  |      |                |        

It provides a fast `Logger` struct type and a simple-to-use `SugaredLogger`
wrapping the former for non-time-critical code. Both embed a `Core` interface,
the implementations of which wrap a `WriteSyncer` interface 
(`io.Writer` + `func sync() error`), which is notably implemented by `os.File`.

The `SugaredLogger` provides per-level methods inspired by `Print`, `Println` and
`Printf`.


- The package does not include a readily injectable interface.
- Encoding: it includes a fast, zero-allocation, JSON encoder (`zapcore.jsonEncoder`), 
  which it uses by default (overridable) to encode logs.
- Stack: it allows adding a stack trace to logs  
- This is the package used in the PoC.
- Stats: 7k dependents, 10k stars, 700 forks (4 years)


### sirupsen/logrus

As the original high-caliber structured/leveled logger for Go, Logrus is the most
popular package on https://pkg.go.dev .

It is now entirely stable with no evolution planned beyond corrective maintenance,
and recommends that new projects should rather use a more recent design, including
apex/log, rs/zerolog, and uber/zap.

It is compatible with the runtime `log` package, allowing it to be used as
a replacement for the default logger by passing the result of its `Writer()`
method to `log.SetOutput`.

Logrus defines 7 logging levels, which do not really map to RFC5424:

| RFC5424 | Name                        |Logrus| Logrus name    |
|:-------:|-----------------------------|:----:|----------------|
|         |                             | 6    | Trace          |
| 7       | Debug                       | 5    | Debug          |       
| 6       | Informational               | 4    | Info           |
| 5       | Notice                      |      |                |
| 4       | Warning                     | 3    | Warning        |
| 3       | Error                       | 2    | Error          |
| 2       | Critical                    |      |                |
| 1       | Alert                       | 1    | Fatal          |
| 0       | Emergency: system unusable  | 0    | Panic          |        


It favors field-based logs rather than plain messages.
  
- The package includes 2 injectable interfaces but recommends its concrete `Logger`
  type for parameters instead.
- Stack: it allows adding a stack trace to logs (`setReportCaller()`)  
- Stats: 27k dependents, 15k stars, 1700 forks, 6.5 years old.

Based on the above, it is not an optimal choice.


### apex/log

This is the structured/leveled logger designed by TJ Holowaychuck, of Drupal and
Node.js (Express) fame. It is inspired by sirupsen/logrus.

https://medium.com/@tjholowaychuk/apex-log-e8d9627f4a9a#.rav8yhkud

Apex/log defines 6 logging levels, which do not really map to RFC5424:

| RFC5424 | Name                        | Log  | apex/log name  |
|:-------:|-----------------------------|:----:|----------------|
|         |                             | -1   | Invalid        |
| 7       | Debug                       | 0    | Debug          |       
| 6       | Informational               | 1    | Info           |
| 5       | Notice                      |      |                |
| 4       | Warning                     | 2    | Warning        |
| 3       | Error                       | 3    | Error          |
| 2       | Critical                    |      |                |
| 1       | Alert                       | 4    | Fatal          |
| 0       | Emergency: system unusable  |      |                |        

- Its API is closely derived from sirupsen/logrus. However, unlike logrus, it 
  does not include a `Writer()` method allowing it to be used as a log destination 
  for the runtime `log` package.
- It is more allocation-heavy than even logrus, many times over uber/zap and rs/zerolog  
- The package provides an injectable `log.Interface` interface and a default
  implementation.
- Stack: it will ship a stack trace on errors implementing `func StackTrace() errors.StackTrace`  


### rs/zerolog

This is a structured/leveled logger created by Netflix, inspired by uber/zap,
aiming for a simpler API and similar or better performance. 

rs/zerolog defines 7 logging levels, which do not really map to RFC5424:

| RFC5424 | Name                        |zerolog|rs/zerolog name|
|:-------:|-----------------------------|:----:|----------------|
|         |                             | -1   | Trace          |
| 7       | Debug                       | 0    | Debug          |       
| 6       | Informational               | 1    | Info           |
| 5       | Notice                      |      |                |
| 4       | Warning                     | 2    | Warning        |
| 3       | Error                       | 3    | Error          |
| 2       | Critical                    |      |                |
| 1       | Alert                       | 4    | Fatal          |
| 0       | Emergency: system unusable  | 5    | Panic          |
|         |                             | 6    | No level (not sent) |
|         |                             | 7    | Disabled logger (not sent) |        

- Its API is closely derived from sirupsen/logrus. However, unlike logrus, it 
  does not include a `Writer()` method allowing it to be used as a log destination 
  for the runtime `log` package.
- Level and message are optional.
- Stack: it can log stack traces
- A RFC7049 CBOR encoding is available for better performance than JSON.
- The package does not provides an injectable specific interface, but its concrete
  `Logger` type implements `io.Writer`, making it immediately compatibly with the
  runtime `log` package.
- Stats: 1k dependents, 3k stars, 200 forks (3 years).  

With performance, `log` compatibility, and stacktrace ability, this is our choice.
