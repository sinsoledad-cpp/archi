package logger

import (
	"strings"
	"time"
)

var DEBUG = false

func Error(err error) Field {
	return Field{Key: "error", Val: err}
}
func Debug(msg string) Field {
	return Field{Key: "debug", Val: msg}
}
func Info(msg string) Field {
	return Field{Key: "info", Val: msg}
}
func Warn(msg string) Field {
	return Field{Key: "warn", Val: msg}
}

func SafeString(key, val string) Field {
	if DEBUG {
		return Field{Key: key, Val: val}
	} else {
		return Field{Key: key, Val: "***"}
	}
}

// SafePhoneZH 安全地返回中国手机号，eg: 135****1234
func SafePhoneZH(phone string) Field {
	if DEBUG {
		return Field{Key: "phone_zh", Val: phone}
	} else {
		return Field{Key: "phone_zh", Val: phone[:3] + "****" + phone[len(phone)-4:]}
	}
}

// SafeEmail 安全地返回邮箱，eg: ***@example.com
func SafeEmail(email string) Field {
	if DEBUG {
		return Field{Key: "email", Val: email}
	} else {
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			panic("Invalid email")
		}
		return Field{Key: "email", Val: "@" + parts[1]}
	}
}

func Any[T any](key string, val T) Field {
	return Field{Key: key, Val: val}
}

func Slice[T any](key string, slice []T) Field {
	return Field{Key: key, Val: slice}
}

func String(key, val string) Field {
	return Field{Key: key, Val: val}
}

func Bool(key string, val bool) Field {
	return Field{Key: key, Val: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Val: val}
}

func Int32(key string, val int32) Field {
	return Field{Key: key, Val: val}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Val: val}
}

// TimeString 默认 layout 为 time.RFC3339，仅第一个有效
func TimeString(val time.Time, layout ...string) Field {
	if len(layout) > 0 {
		return Field{Key: "time", Val: val.Format(layout[0])}
	}
	return Field{Key: "time", Val: val.Format(time.RFC3339)}
}

func TimeUnixMilli(val time.Time) Field {
	return Field{Key: "time", Val: val.UnixMilli()}
}

func TimeUnixNano(val time.Time) Field {
	return Field{Key: "time", Val: val.UnixNano()}
}

func TimeUnixMicros(val time.Time) Field {
	return Field{Key: "time", Val: val.UnixMicro()}
}

func TimeUnix(val time.Time) Field {
	return Field{Key: "time", Val: val.Unix()}
}
