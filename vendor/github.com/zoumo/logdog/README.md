# Logdog

# Why logdog

> Log is the most loyal friend for programmers
> -- Jim Zhang

Logdog is a Go logging package for Humans™ inspired by Python’s logging module. 

# Overview
![terminal.png](http://7xjgzy.com1.z0.glb.clouddn.com/logdog/terminal.png)

# Getting Started

```go
package main

import (
	"github.com/zoumo/logdog"
)

func init() {
	// set min level, only log info or above can be emitted
	logdog.SetLevel(logdog.INFO)

	// enable logger get caller func's filename and lineno
	logdog.EnableRuntimeCaller(true)
}

func main() {
	logdog.Debug("this is debug")
	// Fields should be the last arg
	logdog.Info("this is info", logdog.Fields{"x": "test"})
}
```

# Introduce

## Logging Flow
![logging flow](http://7xjgzy.com1.z0.glb.clouddn.com/logdog/logging_flow.png)

## Logger Mindmap
![logger mindmap](http://7xjgzy.com1.z0.glb.clouddn.com/logdog/logger_mindnode.png)

## Fields
Inspired by [Logrus](https://github.com/Sirupsen/logrus).
the Fields must be the **LAST** arg in log fucntion.

```go
	logdog.Info("this is info, msg ", "some msg", logdog.Fields{"x": "test"})
	logdog.Infof("this is info, msg %s", "some msg", logdog.Fields{"x": "test"})
```

## Loggers
`Logger` have a threefold job. 
First, they expose several methods to application code so that applications can log messages at runtime. 
Second, logger objects determine which log messages to act upon based upon severity (the default filtering facility) or filter objects. 
Third, logger objects pass along relevant log messages to all interested log handlers.

> I do not adopt the inheritance features in Python logging, because it is obscure, intricate and useless. I would like the Logger be simple and readable

## Handlers
`Handler` is responsible for dispatching the appropriate log messages to the handler’s specified destination. 
`Logger` can add zero or more handlers to themselves with `AddHandler()` method. For example to send all log messages to stdout , all log messages of error or higher level to a log file

```go
import (
    "github.com/zoumo/logdog"
)

func main() {
    logdog.DisableExistingLoggers()
    handler := logdog.NewStreamHandler()
    handler2 := logdog.NewFileHandler()
    handler2.LoadConfig(logdog.Config{
        "level"   : "ERROR",
		"filename": "./test",
	})
    logdog.AddHandler(handler)
    
    logdog.Debug("this is debug")
    logdog.Error("this is error")
}
```

`Handler` is a _Interface Type_. 

```go
type Handler interface {
    // Handle the specified record, filter and emit
    Handle(*LogRecord)
    // Check if handler should filter the specified record
    Filter(*LogRecord) bool
    // Emit log record to output - e.g. stderr or file
    Emit(*LogRecord)
    // Close output stream, if not return error
    Close() error
}
```

Logdog comes with built-in handlers: `NullHandler`, `SteamHandler`, `FileHandler`. Maybe provide more extra handlers in the future, e.g. `RotatingFileHandler`

## Formatters
`Formatters` configure the final order, structure, and contents of the log message
Each `Handler` contains one `Formatter`, because only `Handler` itself knows which `Formatter` should be selected to determine the order, structure, and contents of log message
Logdog comes with built-in formatters: `TextFormatter`, `JsonFormatter`
`Formatter` is a _Interface Type_

```go
type Formatter interface {
	Format(*LogRecord) (string, error)
}
```

### TextFormatter
the default `TextFormatter` takes three args: 

| arg          | description                        | default |
| ------------ | ---------------------------------- | ------- |
| DateFmt      | date time format string            | "%Y-%m-%d %H:%M:%S" |
| Fmt          | log message format string          | %(color)[%(time)] [%(levelname)] [%(filename):%(lineno)]%(end_color) %(message) |
| EnableColors | enable print log with color or not | true    |

The **DateFmt** format string looks like python datetime format string
the possible keys  are documented in [go-when Strftime](https://github.com/zoumo/go-when#strftime)

The **Fmt** message format string uses `%(<dictionary key>)` styled string substitution; the possible keys :

| key name       | description                              |
| -------------- | ---------------------------------------- |
| name           | Name of the logger (logging channel)     |
| levelno        | Numeric logging level for the message (DEBUG, INFO, WARNING, ERROR, CRITICAL) |
| levelname      | Text logging level for the message ("DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL") |
| pathname       | Full pathname of the source file where the logging call was issued (if available) or maybe ?? |
| filename       | Filename portion of pathname             |
| lineno         | Source line number where the logging call was issued (if available) |
| funcname       | Function name of caller or maybe ??      |
| time           | Textual time when the LogRecord was created |
| message        | The result of record.getMessage(), computed just as the record is emitted |
| color          | print color                              |
| end_color      | reset color                              |

# Configuring Logging
Programmers can configure logging in two ways:

1. Creating loggers, handlers, and formatters explicitly using Golang code that calls the configuration methods listed above.
2. Creating a json format config string and passing it to the LoadJSONConfig() function.

```go
package main

import (
	"github.com/zoumo/logdog"
)

func main() {
    
    logger := logdog.GetLogger("test")
    
    handler := logdog.NewStreamHandler()
    handler2 := logdog.NewFileHandler()
    handler2.LoadConfig(logdog.Config{
		"filename": "./test",
    })
    
    logger.AddHandler(handler, handler2)
    
    logger.SetLevel(logdog.INFO)
    logger.EnableRuntimeCaller(true)
    
	logger.Debug("this is debug")
	logger.Info("this is info", logdog.Fields{"x": "test"})
}
```

```go
package main

import (
	"github.com/zoumo/logdog"
)

func main() {

	config := []byte(`{
        "disable_existing_loggers": true,
        "handlers": {
            "null": {
                "class": "NullHandler",
                "level": "DEBUG"
            },
            "console": {
                "class": "StreamHandler",
                "formatter": "default",
                "level": "INFO"
            }
        },
        "loggers": {
            "app": {
                "level": "DEBUG",
                "enable_runtime_caller": true,
                "handlers": ["null", "console"]
            }
        }

    }`)

    logdog.LoadJSONConfig(config)
    logger := logdog.GetLogger("app")
    
	logger.Debug("this is debug")
	logger.Info("this is info", logdog.Fields{"x": "test"})
}
```

# Requirement
- [golang.org/x/crypto/ssh/terminal](https://github.com/golang/crypto/tree/master/ssh/terminal)
- [github.com/stretchr/testify/assert](https://github.com/stretchr/testify/assert)

# TODO
- rotating file handler
- use goroutine to accelerate logdog
- more handler
- godoc

# Thanks
- [logrus](https://github.com/Sirupsen/logrus)
- [glogger](https://github.com/Xuyuanp/glogger)
- [timeutil](https://github.com/leekchan/timeutil)