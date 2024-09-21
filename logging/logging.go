/*
Ciphect, a personal data relay.
Copyright (C) 2024 Naeim Cragwell-Chaudhry

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Level string

const (
	Off           Level = "off"
	Fatal         Level = "fatal"
	Critical      Level = "critical"
	Error         Level = "error"
	Warning       Level = "warning"
	Notification  Level = "notification"
	Information   Level = "information"
	Miscellaneous Level = "miscellaneous"
	Debug         Level = "debug"
	Trace         Level = "trace"
	All           Level = "all"
)

func (level Level) Ordinal() int {
	for ord, lvl := range Levels {
		if level == lvl {
			return ord
		}
	}
	return -1
}

var Levels = []Level{Off, Fatal, Critical, Error, Warning, Notification, Information, Miscellaneous, Debug, Trace, All}

type Log struct {
	Timestamp time.Time `json:"timestamp"`
	Level     Level     `json:"level"`
	Log       any       `json:"log"`
}

type Logger struct {
	Level    Level
	Printers []Printer

	c     chan Log
	close chan any
}

func New(level Level, buffer int, printers ...Printer) *Logger {

	logger := &Logger{
		Level:    level,
		Printers: printers,
		c:        make(chan Log, buffer),
		close:    make(chan any),
	}

	go func() {
		for log := range logger.c {
			for _, printer := range logger.Printers {
				printer.Print(log)
			}
		}
		close(logger.close)
	}()

	return logger
}

func (logger *Logger) Close() {
	close(logger.c)
	<-logger.close
}

func (logger *Logger) Printf(level Level, format string, a ...any) {
	logger.Print(level, fmt.Sprintf(format, a...))
}

func (logger *Logger) Print(level Level, a any) {
	if logger.Level.Ordinal() >= level.Ordinal() {
		logger.c <- Log{
			Timestamp: time.Now().UTC(),
			Level:     level,
			Log:       a,
		}
	}
}

type Printer interface {
	Print(Log) error
}

type String struct{ Writer io.Writer }

func (printer String) Print(log Log) (err error) {
	_, err = printer.Writer.Write([]byte(
		fmt.Sprintf(
			"[%s] (%s) %s\n",
			log.Timestamp.Format("2006-01-02 15:04:05 MST"),
			log.Level,
			log.Log,
		),
	))
	return
}

type JSON struct{ Writer io.Writer }

func (printer JSON) Print(log Log) (err error) {
	b, err := json.Marshal(log)
	if err != nil {
		return
	}
	_, err = printer.Writer.Write(append(b, '\n'))
	return
}
