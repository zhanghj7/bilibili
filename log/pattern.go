package log

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var patternMap = map[string]func(map[string]interface{}) string{
	"T": longTime,
	"t": shortTime,
	"D": longDate,
	"d": shortDate,
	"L": keyFactory(_level),
	"f": keyFactory(_source),
	"i": keyFactory(_instanceID),
	"e": keyFactory(_deplyEnv),
	"z": keyFactory(_zone),
	"S": longSource,
	"s": shortSource,
	"M": message,
}

// newPatternRender new pattern render
func newPatternRender(format string) Render {
	p := &pattern{
		bufPool: sync.Pool{New: func() interface{} { return &bytes.Buffer{} }},
	}
	b := make([]byte, 0, len(format))
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			b = append(b, format[i])
			continue
		}
		if i+1 >= len(format) {
			b = append(b, format[i])
			continue
		}
		f, ok := patternMap[string(format[i+1])]
		if !ok {
			b = append(b, format[i])
			continue
		}
		if len(b) != 0 {
			p.funcs = append(p.funcs, textFactory(string(b)))
			b = b[:0]
		}
		p.funcs = append(p.funcs, f)
		i++
	}
	if len(b) != 0 {
		p.funcs = append(p.funcs, textFactory(string(b)))
	}
	return p
}

type pattern struct {
	funcs   []func(map[string]interface{}) string
	bufPool sync.Pool
}

// Render implemet Formater
func (p *pattern) Render(w io.Writer, d map[string]interface{}) error {
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	for _, f := range p.funcs {
		buf.WriteString(f(d))
	}

	_, err := buf.WriteTo(w)
	return err
}

// Render implemet Formater as string
func (p *pattern) RenderString(d map[string]interface{}) string {
	// TODO strings.Builder
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	for _, f := range p.funcs {
		buf.WriteString(f(d))
	}

	return buf.String()
}

func textFactory(text string) func(map[string]interface{}) string {
	return func(map[string]interface{}) string {
		return text
	}
}
func keyFactory(key string) func(map[string]interface{}) string {
	return func(d map[string]interface{}) string {
		if d["colorformat"] != nil {
			colorFormat := d["colorformat"].(string)
			if v, ok := d[key]; ok {
				if s, ok := v.(string); ok {
					return formatColor(colorFormat, s)
				}
				return formatColor(colorFormat, fmt.Sprint(v))
			}
			return ""
		} else {
			if v, ok := d[key]; ok {
				if s, ok := v.(string); ok {
					return s
				}
				return fmt.Sprint(v)
			}
			return ""
		}
	}
}

func longSource(map[string]interface{}) string {
	if pc, file, lineNo, ok := runtime.Caller(5); ok {
		funcName1 := runtime.FuncForPC(pc).Name()    // main.(*MyStruct).foo
		funcName := filepath.Ext(funcName1)          // .foo
		funcName = strings.TrimPrefix(funcName, ".") // fo
		file = funcName1[0:strings.Index(funcName1, ".")] + "/" + filepath.Base(file)
		name := file + ":" + funcName + ":" + strconv.FormatInt(int64(lineNo), 10)
		return name
	}
	//if _, file, lineNo, ok := runtime.Caller(6); ok {
	//	return fmt.Sprintf("%s:%d", file, lineNo)
	//}
	return "unknown:0"
}

func shortSource(map[string]interface{}) string {
	if _, file, lineNo, ok := runtime.Caller(6); ok {
		return fmt.Sprintf("%s:%d", path.Base(file), lineNo)
	}
	return "unknown:0"
}

func longTime(data map[string]interface{}) string {
	date := time.Now().Format("15:04:05.000")
	if data["colorformat"] != nil {
		colorFormat := data["colorformat"].(string)
		date = formatColor(colorFormat, date)
	}
	return date
}

func shortTime(data map[string]interface{}) string {
	date := time.Now().Format("15:04")
	if data["colorformat"] != nil {
		colorFormat := data["colorformat"].(string)
		date = formatColor(colorFormat, date)
	}
	return date
}

func longDate(data map[string]interface{}) string {
	date := time.Now().Format("2006/01/02")
	if data["colorformat"] != nil {
		colorFormat := data["colorformat"].(string)
		date = formatColor(colorFormat, date)
	}
	return date
}

func shortDate(data map[string]interface{}) string {
	date := time.Now().Format("01/02")
	if data["colorformat"] != nil {
		colorFormat := data["colorformat"].(string)
		date = formatColor(colorFormat, date)
	}
	return date
	//return time.Now().Format("01/02")
}

func isInternalKey(k string) bool {
	switch k {
	case _level, _levelValue, _time, _source, _instanceID, _appID, _deplyEnv, _zone, _color_format:
		return true
	}
	return false
}

func message(d map[string]interface{}) string {
	var m string
	var s []string
	for k, v := range d {
		if k == _log {
			m = fmt.Sprint(v)
			continue
		}
		if isInternalKey(k) {
			continue
		}
		s = append(s, fmt.Sprintf("%s=%v", k, v))
	}
	s = append(s, m)
	s1 := strings.Join(s, " ")
	if d["colorformat"] != nil {
		colorFormat := d["colorformat"].(string)
		s1 = formatColor(colorFormat, s1)
	}
	return s1
}
