package util

import (
	"fmt"
	"runtime"
)

type StackFrame struct {
	Source string `json:"source"`
	Func   string `json:"func"`
}

func GetStack(skip int) []StackFrame {
	var pcs [100]uintptr
	n := runtime.Callers(2+skip, pcs[:])
	framesIter := runtime.CallersFrames(pcs[:n])
	frames := make([]runtime.Frame, 0, n)
	for {
		frame, more := framesIter.Next()
		frames = append(frames, frame)
		if !more {
			break
		}
	}
	stack := make([]StackFrame, len(frames))
	for i, frame := range frames {
		stack[i].Func = frame.Function
		stack[i].Source = fmt.Sprintf("%s:%d", frame.File, frame.Line)
	}
	return stack
}
