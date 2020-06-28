package gox

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

type MarkupOrChild interface {
	isMarkupOrChild()
}

func markupOrChild(i interface{}) MarkupOrChild {
	if m, ok := i.(MarkupOrChild); ok {
		return m
	}
	if c, ok := i.(Component); ok {
		return c.Render()
	}

	switch t := i.(type) {
	default:
		panic(fmt.Errorf("markupOrChild: unsupported type %T", t))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return &HTML{
			text: fmt.Sprintf("%d", t),
		}
	case []ComponentOrHTML:
		return ComponentOrHTMLList(t)
	case string:
		return &HTML{
			text: template.HTMLEscapeString(t),
		}
	}
}

type Component interface {
	Render() ComponentOrHTML
}

type component struct {
	Component
}

func NewComponent(c Component) ComponentOrHTML {
	return &component{
		Component: c,
	}
}

func (component) isMarkupOrChild()   {}
func (component) isComponentOrHTML() {}

func (c component) Render() string {

	c2 := c.Component.Render()

	return c2.Render()
}

type ComponentOrHTML interface {
	isComponentOrHTML()
	isMarkupOrChild()

	Render() string
}

type ComponentOrHTMLList []ComponentOrHTML

func (ComponentOrHTMLList) isMarkupOrChild() {}

type HTML struct {
	tag        string
	text       string
	children   []ComponentOrHTML
	properties map[string]interface{}
}

func (HTML) isMarkupOrChild()   {}
func (HTML) isComponentOrHTML() {}

func (h HTML) Render() string {
	// TODO: handle self-closing tags properly (i.e img)

	b := &strings.Builder{}
	tag := h.tag
	if tag == "" && len(h.properties) != 0 {
		tag = "div"
	}
	if tag != "" && len(h.properties) == 0 {
		fmt.Fprintf(b, "<%s>", tag)
	}
	if tag != "" && len(h.properties) != 0 {
		fmt.Fprintf(b, "<%s ", tag)
		h.renderProperties(b)
		fmt.Fprint(b, "> ")
	}

	if h.text != "" {
		fmt.Fprintf(b, "%s", h.text)
	}

	h.renderChildren(b)

	if tag != "" {
		fmt.Fprintf(b, "</%s>", tag)
	}

	return b.String()
}

func (h HTML) renderChildren(w io.Writer) {
	for _, c := range h.children {
		_, _ = io.WriteString(w, c.Render())
	}
}

func (h HTML) renderProperties(w io.Writer) {
	ss := []string{}
	for k, v := range h.properties {
		ss = append(ss, fmt.Sprintf("%s=%q", k, v))
	}
	fmt.Fprint(w, strings.Join(ss, " "))
}

type MarkupList struct {
	list []Applyer
}

func (MarkupList) isMarkupOrChild() {}

func (m MarkupList) Apply(h *HTML) {
	for _, a := range m.list {
		if a == nil {
			continue
		}
		a.Apply(h)
	}
}

type Applyer interface {
	Apply(h *HTML)
}

func Tag(tag string, mm ...interface{}) *HTML {
	h := &HTML{
		tag: tag,
	}
	for _, m := range mm {
		apply(markupOrChild(m), h)
	}
	return h
}

func Text(text string, mm ...interface{}) *HTML {
	h := &HTML{
		text: template.HTMLEscapeString(text),
	}
	for _, m := range mm {
		apply(markupOrChild(m), h)
	}
	return h
}

func Value(val interface{}) *HTML {
	switch t := val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return &HTML{
			text: fmt.Sprintf("%d", t),
		}
	case string:
		h := &HTML{
			text: template.HTMLEscapeString(t),
		}

		return h
	case Raw:
		return &HTML{
			text: string(t),
		}
	case *HTML:
		return t
	case []ComponentOrHTML:
		return &HTML{
			children: t,
		}
	default:
		panic(fmt.Errorf("Value: unsupported type %T", t))
	}
}

type markupFunc func(h *HTML)

func (m markupFunc) Apply(h *HTML) { m(h) }

func Markup(m ...Applyer) MarkupList {
	return MarkupList{list: m}
}

type Raw string

func (Raw) isMarkupOrChild()   {}
func (Raw) isComponentOrHTML() {}

func (r Raw) Render() string {
	return string(r)
}

func Property(key string, value interface{}) Applyer {
	// if key == "style" {
	// 	panic(errors.New(`gox: Property called with key "style"; style package or Style should be used instead`))
	// }
	return markupFunc(func(h *HTML) {
		if h.properties == nil {
			h.properties = make(map[string]interface{})
		}
		h.properties[key] = value
	})
}

func apply(m MarkupOrChild, h *HTML) {
	switch m := m.(type) {
	case MarkupList:
		m.Apply(h)
	case nil:
		h.children = append(h.children, nil)
	case *HTML, *component:
		h.children = append(h.children, m.(ComponentOrHTML))
	case ComponentOrHTMLList:
		h.children = append(h.children, m...)
	default:
		panic(fmt.Errorf("gox: internal error (unexpected MarkupOrChild type %T)", m))
	}
}
