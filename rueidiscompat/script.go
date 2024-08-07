package rueidiscompat

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

type Scripter interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd
	EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd
	EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd
	ScriptExists(ctx context.Context, hashes ...string) *BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *StringCmd
}

var (
	_ Scripter = (*Compat)(nil)
)

type Script struct {
	src, hash string
}

func NewScript(src string) *Script {
	h := sha1.New()
	_, _ = io.WriteString(h, src)
	return &Script{
		src:  src,
		hash: hex.EncodeToString(h.Sum(nil)),
	}
}

func (s *Script) Hash() string {
	return s.hash
}

func (s *Script) Load(ctx context.Context, c Scripter) *StringCmd {
	return c.ScriptLoad(ctx, s.src)
}

func (s *Script) Exists(ctx context.Context, c Scripter) *BoolSliceCmd {
	return c.ScriptExists(ctx, s.hash)
}

func (s *Script) Eval(ctx context.Context, c Scripter, keys []string, args ...interface{}) *Cmd {
	return c.Eval(ctx, s.src, keys, args...)
}

func (s *Script) EvalRO(ctx context.Context, c Scripter, keys []string, args ...interface{}) *Cmd {
	return c.EvalRO(ctx, s.src, keys, args...)
}

func (s *Script) EvalSha(ctx context.Context, c Scripter, keys []string, args ...interface{}) *Cmd {
	fmt.Println(ctx, s.hash)
	fmt.Println(keys, args)
	return c.EvalSha(ctx, s.hash, keys, args...)
}

func (s *Script) EvalShaRO(ctx context.Context, c Scripter, keys []string, args ...interface{}) *Cmd {
	return c.EvalShaRO(ctx, s.hash, keys, args...)
}

// Run optimistically uses EVALSHA to run the script. If script does not exist
// it is retried using EVAL.
func (s *Script) Run(ctx context.Context, c Scripter, keys []string, args ...interface{}) *Cmd {
	r := s.EvalSha(ctx, c, keys, args...)
	if err := r.Err(); err != nil {
		msg := err.Error()
		msg = strings.TrimPrefix(msg, "ERR ")
		if strings.HasPrefix(msg, "NOSCRIPT") {
			return s.Eval(ctx, c, keys, args...)
		}
	}
	return r
}

// RunRO optimistically uses EVALSHA_RO to run the script. If script does not exist
// it is retried using EVAL_RO.
func (s *Script) RunRO(ctx context.Context, c Scripter, keys []string, args ...interface{}) *Cmd {
	r := s.EvalShaRO(ctx, c, keys, args...)
	if err := r.Err(); err != nil {
		msg := err.Error()
		msg = strings.TrimPrefix(msg, "ERR ")
		if strings.HasPrefix(msg, "NOSCRIPT") {
			return s.Eval(ctx, c, keys, args...)
		}
	}
	return r
}
