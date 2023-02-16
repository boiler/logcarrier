package logging

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	nocolor   = 0
	red       = 31
	green     = 32
	yellow    = 33
	blue      = 34
	lightBlue = 36
)

var (
	baseTimestamp time.Time
	isTerminal    bool
	noQuoteNeeded *regexp.Regexp
)

func init() {
	baseTimestamp = time.Now()
	isTerminal = logrus.IsTerminal(os.Stdout)
}

// This is to not silently overwrite `time`, `msg` and `level` fields when
// dumping it. If this code wasn't there doing:
//
// logrus.WithField("level", 1).Info("hello")
//
// Would just silently drop the user provided level. Instead with this code
// it'll logged as:
//
// {"level": "info", "fields.level": 1, "msg": "hello", "time": "..."}
//
// It's not exported because it's still using Data in an opinionated way. It's to
// avoid code duplication between the two default formatters.
func prefixFieldClashes(data logrus.Fields) {
	_, ok := data["time"]
	if ok {
		data["fields.time"] = data["time"]
	}
	_, ok = data["msg"]
	if ok {
		data["fields.msg"] = data["msg"]
	}
	_, ok = data["level"]
	if ok {
		data["fields.level"] = data["level"]
	}
}

// TextFormatter копипаста logrus.TextFormatter с косметическими изменениями
type TextFormatter struct {
	// // Set to true to bypass checking for a TTY before outputting colors.
	// ForceColors   bool
	// DisableColors bool
	// // Set to true to disable timestamp logging (useful when the output
	// // is redirected to a logging system already adding a timestamp)
	// DisableTimestamp bool
}

// Format возвращает текст ошибки
func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	// pp.Println(entry)

	var keys []string
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b := &bytes.Buffer{}

	prefixFieldClashes(entry.Data)

	// isColored := (f.ForceColors || isTerminal) && !f.DisableColors
	isColored := entry.Logger.Out == os.Stderr
	// @TODO: Check
	// Out: &os.File{
	//   file: &os.file{
	//     fd:      2,
	//     name:    "/dev/stderr",
	//     dirinfo: (*os.dirInfo)(nil),
	//     nepipe:  0,
	//   },

	var levelColor int

	messagePrefix := fmt.Sprintf("%s ", strings.ToUpper(entry.Level.String())[0:1])

	if strings.HasPrefix(entry.Message, "C ") { // hack for critical. Не добавлять префикс если он уже есть
		messagePrefix = ""
	}

	if isColored {
		switch entry.Level {
		case logrus.WarnLevel:
			levelColor = yellow
		case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			levelColor = red
		case logrus.DebugLevel:
			levelColor = lightBlue
		default:
			levelColor = blue
		}

		fmt.Fprintf(b, "\x1b[%dm[%s] %s\x1b[0m%s",
			levelColor,
			entry.Time.Format("2006-01-02 15:04:05"),
			messagePrefix,
			entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] %s%s",
			entry.Time.Format("2006-01-02 15:04:05"),
			messagePrefix,
			entry.Message)
	}

	if len(keys) > 0 {
		b.WriteString(" {")
		if isColored {
			for index, k := range keys {
				if index > 0 {
					b.WriteString(", ")
				}
				v := entry.Data[k]
				fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m=%#v", levelColor, k, v)
			}
		} else {
			for index, k := range keys {
				if index > 0 {
					b.WriteString(", ")
				}
				v := entry.Data[k]
				fmt.Fprintf(b, "%s=%#v", k, v)
			}
		}
		b.WriteString("}")
	}
	b.WriteString("\n")

	return b.Bytes(), nil
}

func needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch < '9') ||
			ch == '-' || ch == '.') {
			return false
		}
	}
	return true
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key, value interface{}) {
	switch value.(type) {
	case string:
		if needsQuoting(value.(string)) {
			fmt.Fprintf(b, "%v=%s ", key, value)
		} else {
			fmt.Fprintf(b, "%v=%q ", key, value)
		}
	case error:
		if needsQuoting(value.(error).Error()) {
			fmt.Fprintf(b, "%v=%s ", key, value)
		} else {
			fmt.Fprintf(b, "%v=%q ", key, value)
		}
	default:
		fmt.Fprintf(b, "%v=%v ", key, value)
	}
}
