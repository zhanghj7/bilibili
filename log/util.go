package log

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/zhanghj7/bilibili/conf/env"
	"github.com/zhanghj7/bilibili/net/metadata"
	"github.com/zhanghj7/bilibili/net/trace"
)

var fm sync.Map

func addExtraField(ctx context.Context, fields map[string]interface{}) {
	if t, ok := trace.FromContext(ctx); ok {
		if s, ok := t.(fmt.Stringer); ok {
			fields[_tid] = s.String()
		} else {
			fields[_tid] = fmt.Sprintf("%s", t)
		}
	}
	if caller := metadata.String(ctx, metadata.Caller); caller != "" {
		fields[_caller] = caller
	}
	if color := metadata.String(ctx, metadata.Color); color != "" {
		fields[_color] = color
	}
	if cluster := metadata.String(ctx, metadata.Cluster); cluster != "" {
		fields[_cluster] = cluster
	}
	fields[_deplyEnv] = env.DeployEnv
	fields[_zone] = env.Zone
	fields[_appID] = c.Family
	fields[_instanceID] = c.Host
	if metadata.Bool(ctx, metadata.Mirror) {
		fields[_mirror] = true
	}
}

// funcName get func name.
func funcName(skip int) (name string) {
	if pc, file, lineNo, ok := runtime.Caller(skip); ok {
		if v, ok := fm.Load(pc); ok {
			name = v.(string)
		} else {
			//name = runtime.FuncForPC(pc).Name() + ":" + strconv.FormatInt(int64(lineNo), 10)
			funcname1 := runtime.FuncForPC(pc).Name()    // main.(*MyStruct).foo
			funcname := filepath.Ext(funcname1)          // .foo
			funcname = strings.TrimPrefix(funcname, ".") // fo
			file = funcname1[0:strings.Index(funcname1, ".")] + "/" + filepath.Base(file)
			name = file + ":" + funcname + ":" + strconv.FormatInt(int64(lineNo), 10)
			fm.Store(pc, name)
		}
	}
	return
}
