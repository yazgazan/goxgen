package gox

import (
	"fmt"
	"html/template"
	"io"
	"strings"
)

type MarkupOrChild interface {
	isMarkupOrChild()
}

type Writer func(io.Writer) error
type HTML = Writer

func (w Writer) isMarkupOrChild() {}
func (w Writer) String() (string, error) {
	b := &strings.Builder{}
	err := w(b)

	return b.String(), err
}
func (w Writer) WriteTo(dst io.Writer) error {
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

func Tag(tag string, mm ...MarkupOrChild) Writer {
	return func(w io.Writer) error {
		markup, children, err := markupAndChildren(mm)
		if err != nil {
			return err
		}

		props := make(map[string]interface{}, len(markup))
		markup.Apply(props)

		if tag == "" && len(props) != 0 {
			tag = "div"
		}
		if tag != "" && len(props) == 0 {
			_, err := fmt.Fprintf(w, "<%s>", tag)
			if err != nil {
				return err
			}
		}
		if tag != "" && len(props) != 0 {
			_, err := fmt.Fprintf(w, "<%s", tag)
			if err != nil {
				return err
			}
			err = renderProperties(w, props)
			if err != nil {
				return err
			}
			_, err = w.Write([]byte(">"))
			if err != nil {
				return err
			}
		}

		for _, wr := range children {
			err := wr(w)
			if err != nil {
				return err
			}
		}

		if tag != "" {
			_, err := fmt.Fprintf(w, "</%s>", tag)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func Text(text string) Writer {
	return func(w io.Writer) error {
		_, err := io.WriteString(w, template.HTMLEscapeString(text))

		return err
	}
}

func Writers(cc ...Writer) Writer {
	return func(w io.Writer) error {
		for _, c := range cc {
			err := c(w)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

type Component interface {
	Render() Writer
}

func NewComponent(c Component) Writer {
	return c.Render()
}

func renderProperties(w io.Writer, props map[string]interface{}) error {
	for k, v := range props {
		_, err := fmt.Fprintf(w, " %s=%q", k, v)
		if err != nil {
			return err
		}
	}

	return nil
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
	return func(w io.Writer) error {
		var err error

		switch t := val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			_, err = fmt.Fprintf(w, "%d", t)
		case string:
			_, err = io.WriteString(w, template.HTMLEscapeString(t))
		case Raw:
			_, err = io.WriteString(w, string(t))
		case Writer:
			err = t(w)
		case []Writer:
			for _, child := range t {
				err = child(w)
				if err != nil {
					break
				}
			}
		default:
			err = fmt.Errorf("Value: unsupported type %T", t)
		}

		return err
	}
}

type Raw string

func (Raw) isMarkupOrChild() {}

func Error(err error) Writer {
	return func(io.Writer) error {
		return err
	}
}
