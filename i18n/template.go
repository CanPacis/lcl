package i18n

import (
	"bytes"
	"context"
	"io"
)

type Template[T any] interface {
	Get(v T) string
}

type Partitioned interface {
	Partitions() []Renderer
}

type Renderer interface {
	Render(context.Context, io.Writer) error
}
type renderFn func(context.Context, io.Writer) error

type renderer struct {
	fn renderFn
}

func (r *renderer) Render(ctx context.Context, w io.Writer) error {
	return r.fn(ctx, w)
}

func renderFunc(fn renderFn) Renderer {
	return &renderer{fn: fn}
}

type errWriter struct {
	w   io.Writer
	n   int
	err error
}

func (w errWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return w.n, w.err
	}

	w.n, w.err = w.w.Write(p)
	return w.n, w.err
}

func (w errWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

type bufferedTemplate[T any] struct {
	renderer Renderer
}

type valueKeyType int

var valueKey valueKeyType = 0

func (t *bufferedTemplate[T]) Get(v T) string {
	ctx := context.WithValue(context.Background(), valueKey, v)
	buf := bytes.NewBuffer([]byte{})
	if err := t.renderer.Render(ctx, buf); err != nil {
		panic(err)
	}
	return buf.String()
}

type StringRenderer string

func (s StringRenderer) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(s))
	return err
}

type partitionedTemplate[T any] struct {
	parititions []Renderer
}

func (t *partitionedTemplate[T]) Get(v T) string {
	ctx := context.WithValue(context.Background(), valueKey, v)
	buf := bytes.NewBuffer([]byte{})

	for _, part := range t.parititions {
		part.Render(ctx, buf)
	}
	return buf.String()
}

func (t *partitionedTemplate[T]) Partitions() []Renderer {
	return t.parititions
}
