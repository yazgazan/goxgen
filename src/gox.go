package gox

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
)

type MarkupOrChild interface {
	isMarkupOrChild()
}

type Writer func(io.Writer) (int64, error)
type HTML = Writer

func (w Writer) isMarkupOrChild() {}
func (w Writer) String() (string, error) {
	b := &bytes.Buffer{}
	_, err := w.WriteTo(b)

	return b.String(), err
}
func (w Writer) WriteTo(dst io.Writer) (int64, error) {
	return w(dst)
}

func markupAndChildren(mm []MarkupOrChild) (MarkupList, []Writer, error) {
	var (
		markup   MarkupList
		children []Writer
	)
	for _, m := range mm {
		if ml, ok := m.(MarkupList); ok {
			markup = append(markup, ml...)
			continue
		}
		if a, ok := m.(Applyer); ok {
			markup = append(markup, a)
			continue
		}
		if wr, ok := m.(Writer); ok {
			children = append(children, wr)
			continue
		}

		return nil, nil, fmt.Errorf("unsupported type %T in Tag()", m)
	}

	return markup, children, nil
}

var (
	selfClosingTags = map[string]bool{
		"area":     true,
		"base":     true,
		"br":       true,
		"col":      true,
		"embed":    true,
		"hr":       true,
		"img":      true,
		"input":    true,
		"link":     true,
		"meta":     true,
		"param":    true,
		"source":   true,
		"track":    true,
		"wbr":      true,
		"command":  true,
		"keygen":   true,
		"menuitem": true,
	}
)

func Doctype(typ string, document Writer) Writer {
	return func(dst io.Writer) (int64, error) {
		n, err := fmt.Fprintf(dst, "<!DOCTYPE %s>\n", typ)
		if err != nil {
			return int64(n), err
		}

		return document.WriteTo(dst)
	}
}

func Tag(tag string, mm ...MarkupOrChild) Writer {
	return func(dst io.Writer) (int64, error) {
		markup, children, err := markupAndChildren(mm)
		if err != nil {
			return 0, err
		}

		props := make(map[string]interface{}, len(markup))
		markup.Apply(props)

		if tag == "" && len(props) != 0 {
			tag = "div"
		}

		selfClosing := selfClosingTags[tag] && len(children) == 0

		n, err := renderOtag(dst, tag, props, selfClosing)
		if err != nil {
			return n, err
		}

		if selfClosing {
			return n, nil
		}

		for _, wr := range children {
			nd, err := wr.WriteTo(dst)
			n += nd
			if err != nil {
				return n, err
			}
		}

		nd, err := renderCtag(dst, tag)
		n += nd
		if err != nil {
			return n, err
		}

		return n, nil
	}
}

func PlainText(ww ...Writer) Writer {
	return func(out io.Writer) (int64, error) {
		var nb int64

		for _, w := range ww {
			n, err := w.WriteTo(out)
			nb += n
			if err != nil {
				return nb, err
			}
		}

		return nb, nil
	}
}

func renderOtag(dst io.Writer, tag string, props map[string]interface{}, selfClosing bool) (int64, error) {
	var n int

	if tag == "" {
		return 0, nil
	}

	if len(props) == 0 && !selfClosing {
		n, err := fmt.Fprintf(dst, "<%s>", tag)
		return int64(n), err
	}
	if len(props) == 0 && selfClosing {
		n, err := fmt.Fprintf(dst, "<%s/>", tag)
		return int64(n), err
	}

	nd, err := fmt.Fprintf(dst, "<%s", tag)
	n += nd
	if err != nil {
		return int64(n), err
	}
	nd, err = renderProperties(dst, props)
	n += nd
	if err != nil {
		return int64(n), err
	}

	if selfClosing {
		nd, err = dst.Write([]byte("/>"))
		n += nd

		return int64(n), err
	}

	nd, err = dst.Write([]byte(">"))
	n += nd

	return int64(n), err
}

func renderCtag(dst io.Writer, tag string) (int64, error) {
	if tag == "" {
		return 0, nil
	}

	n, err := fmt.Fprintf(dst, "</%s>", tag)

	return int64(n), err
}

func Text(text string) Writer {
	return func(dst io.Writer) (int64, error) {
		n, err := io.WriteString(dst, template.HTMLEscapeString(text))

		return int64(n), err
	}
}

func Writers(ww ...Writer) Writer {
	return func(dst io.Writer) (int64, error) {
		var n int64

		for _, w := range ww {
			nd, err := w.WriteTo(dst)
			n += nd
			if err != nil {
				return n, err
			}
		}

		return n, nil
	}
}

type Component interface {
	Render() Writer
}

func NewComponent(c Component) Writer {
	return c.Render()
}

func renderProperties(dst io.Writer, props map[string]interface{}) (int, error) {
	var n int

	for k, v := range props {
		s := fmt.Sprint(v)
		nd, err := fmt.Fprintf(dst, " %s=%q", k, template.HTMLEscapeString(s))
		if err != nil {
			return n + nd, err
		}
		n += nd
	}

	return n, nil
}

type Applyer interface {
	Apply(mm map[string]interface{})
}

type Attributes = Applyer

type MarkupList []Applyer

func (MarkupList) isMarkupOrChild() {}

func (m MarkupList) Apply(mm map[string]interface{}) {
	for _, a := range m {
		if a == nil {
			continue
		}
		a.Apply(mm)
	}
}

func Markup(m ...Applyer) MarkupList {
	return MarkupList(m)
}

type markupFunc func(mm map[string]interface{})

func (m markupFunc) Apply(mm map[string]interface{}) { m(mm) }

func Property(key string, value interface{}) Applyer {
	return markupFunc(func(mm map[string]interface{}) {
		mm[key] = value
	})
}

func Value(val interface{}) Writer {
	return func(dst io.Writer) (int64, error) {
		var (
			err error
			n   int64
			ni  int
		)

		switch t := val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			ni, err = fmt.Fprintf(dst, "%d", t)
		case string:
			ni, err = io.WriteString(dst, template.HTMLEscapeString(t))
		case Raw:
			ni, err = io.WriteString(dst, string(t))
		case Writer:
			n, err = t.WriteTo(dst)
		case []Writer:
			for _, child := range t {
				var (
					n2 int64
					n3 int
				)

				n2, err = child.WriteTo(dst)
				n += n2
				if err != nil {
					break
				}
				n3, err = dst.Write([]byte{'\n'})
				n += int64(n3)
				if err != nil {
					break
				}
			}
		default:
			err = fmt.Errorf("Value: unsupported type %T", t)
		}

		return n + int64(ni), err
	}
}

type Raw string

func (Raw) isMarkupOrChild() {}

func Error(err error) Writer {
	return func(io.Writer) (int64, error) {
		return 0, err
	}
}
