package i18n

import (
	"context"
	"io"
	"time"

	"golang.org/x/text/language"
)

type sec_home_footer struct {
	Label Template[time.Time]
}

type sec_home struct {
	Header         string
	Bio            Template[User]
	PartitionedBio Template[User]
	Footer         sec_home_footer
	Spread         Template[[]string]
}

type I18n struct {
	Home sec_home
}

var Local = map[language.Tag]I18n{
	language.English: {
		Home: sec_home{
			Header: "This is the home header",
			Bio: &bufferedTemplate[User]{
				renderer: renderFunc(func(ctx context.Context, w io.Writer) error {
					buf := errWriter{w: w}
					v := ctx.Value(valueKey).(User)
					buf.WriteString("Name ")
					buf.WriteString(v.Name)
					buf.WriteString(", age ")
					buf.WriteString(strconv_ext_i.Itoa(v.Age))
					return buf.err
				}),
			},
			PartitionedBio: &partitionedTemplate[User]{
				parititions: []Renderer{
					StringRenderer("Name "),
					renderFunc(func(ctx context.Context, w io.Writer) error {
						buf := errWriter{w: w}
						v := ctx.Value(valueKey).(User)
						buf.WriteString(v.Name)
						return buf.err
					}),
					StringRenderer(", age "),
					renderFunc(func(ctx context.Context, w io.Writer) error {
						buf := errWriter{w: w}
						v := ctx.Value(valueKey).(User)
						buf.WriteString(strconv_ext_i.Itoa(v.Age))
						return buf.err
					}),
				},
			},
			Footer: sec_home_footer{
				Label: &bufferedTemplate[time.Time]{
					renderer: renderFunc(func(ctx context.Context, w io.Writer) error {
						buf := errWriter{w: w}
						v := ctx.Value(valueKey).(time.Time)
						buf.WriteString("Since ")
						buf.WriteString(proc_year(v))
						return buf.err
					}),
				},
			},
			Spread: &bufferedTemplate[[]string]{
				renderer: renderFunc(func(ctx context.Context, w io.Writer) error {
					buf := errWriter{w: w}
					v := ctx.Value(valueKey).([]string)
					buf.WriteString(list_ext_en.CommaJoin(v))
					buf.WriteString(" liked your post")
					return buf.err
				}),
			},
		},
	},
	language.Spanish: {},
	language.French:  {},
	language.German:  {},
	language.Turkish: {
		Home: sec_home{
			Header: "...",
			Bio: &bufferedTemplate[User]{
				renderer: renderFunc(func(ctx context.Context, w io.Writer) error {
					buf := errWriter{w: w}
					v := ctx.Value(valueKey).(User)
					buf.WriteString("İsim ")
					buf.WriteString(v.Name)
					buf.WriteString(", yaş ")
					buf.WriteString(strconv_ext_i.Itoa(v.Age))
					return buf.err
				}),
			},
			PartitionedBio: &partitionedTemplate[User]{
				parititions: []Renderer{
					StringRenderer("İsim "),
					renderFunc(func(ctx context.Context, w io.Writer) error {
						buf := errWriter{w: w}
						v := ctx.Value(valueKey).(User)
						buf.WriteString(v.Name)
						return buf.err
					}),
					StringRenderer(", yaş "),
					renderFunc(func(ctx context.Context, w io.Writer) error {
						buf := errWriter{w: w}
						v := ctx.Value(valueKey).(User)
						buf.WriteString(strconv_ext_i.Itoa(v.Age))
						return buf.err
					}),
				},
			},
			Footer: sec_home_footer{
				Label: &bufferedTemplate[time.Time]{
					renderer: renderFunc(func(ctx context.Context, w io.Writer) error {
						buf := errWriter{w: w}
						v := ctx.Value(valueKey).(time.Time)
						buf.WriteString(proc_year(v))
						buf.WriteString("'")
						buf.WriteString(tr_case_ext_i.Ablative(proc_year(v)))
						buf.WriteString(" beri")
						return buf.err
					}),
				},
			},
			Spread: &bufferedTemplate[[]string]{
				renderer: renderFunc(func(ctx context.Context, w io.Writer) error {
					buf := errWriter{w: w}
					v := ctx.Value(valueKey).([]string)
					buf.WriteString(list_ext_tr.CommaJoin(v))
					buf.WriteString(" paylaşımını beğendi")
					return buf.err
				}),
			},
		},
	},
}
