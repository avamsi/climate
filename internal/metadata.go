package internal

import (
	"bytes"
	"encoding/gob"
	"go/ast"
	"reflect"
	"strings"
	"unicode"

	"github.com/avamsi/ergo"
	"github.com/sanity-io/litter"
)

// Note: it's important that all fields of RawMetadata be exported, otherwise
// gob won't be able to encode / decode them correctly.
type RawMetadata struct {
	Doc        string
	Directives map[string]string
	Comment    string
	Params     []string
	Children   map[string]*RawMetadata
}

const directivePrefix = "//climate:"

func (rmd *RawMetadata) SetDoc(doc *ast.CommentGroup) {
	if doc == nil {
		return
	}
	rmd.Doc = strings.TrimSpace(doc.Text())
	rmd.Directives = map[string]string{}
	for _, comment := range doc.List {
		if !strings.HasPrefix(comment.Text, directivePrefix) {
			continue
		}
		d, value, _ := strings.Cut(comment.Text, " ")
		d = strings.TrimPrefix(d, directivePrefix)
		if _, ok := rmd.Directives[d]; ok {
			ergo.Panicf("more than one %q directive: %s", d, litter.Sdump(doc))
		}
		rmd.Directives[d] = strings.TrimSpace(value)
	}
}

func (rmd *RawMetadata) SetComment(comment *ast.CommentGroup) {
	rmd.Comment = strings.TrimSpace(comment.Text())
}

func (rmd *RawMetadata) Child(name string) *RawMetadata {
	if rmd.Children == nil {
		rmd.Children = map[string]*RawMetadata{}
	}
	child, ok := rmd.Children[name]
	if !ok {
		child = &RawMetadata{}
		rmd.Children[name] = child
	}
	return child
}

func (rmd *RawMetadata) Encode() []byte {
	var b bytes.Buffer
	ergo.Must0(gob.NewEncoder(&b).Encode(rmd))
	return b.Bytes()
}

type Metadata struct {
	root     *Metadata
	raw      *RawMetadata
	children map[string]*Metadata
}

func DecodeMetadata(b []byte) *Metadata {
	var raw RawMetadata
	ergo.Must0(gob.NewDecoder(bytes.NewReader(b)).Decode(&raw))
	md := &Metadata{raw: &raw}
	md.root = md
	return md
}

func (md *Metadata) Lookup(pkgPath, name string) *Metadata {
	if md == nil {
		return nil
	}
	return md.root.Child(pkgPath).Child(name)
}

func (md *Metadata) LookupType(t reflect.Type) *Metadata {
	return md.Lookup(t.PkgPath(), t.Name())
}

func (md *Metadata) Aliases() []string {
	if md == nil {
		return nil
	}
	var (
		tmp     = strings.Split(md.raw.Directives["aliases"], ",")
		aliases []string
	)
	for _, alias := range tmp {
		if alias := strings.TrimSpace(alias); alias != "" {
			aliases = append(aliases, alias)
		}
	}
	return aliases
}

func (md *Metadata) Long() string {
	if md == nil {
		return ""
	}
	return md.raw.Doc
}

func (md *Metadata) Short() string {
	if md == nil {
		return ""
	}
	if short, ok := md.raw.Directives["short"]; ok {
		return short
	}
	if md.raw.Comment != "" {
		return md.raw.Comment
	}
	// Auto generate a short description from the long description.
	var (
		long = md.Long()
		i    = strings.Index(long, "\n\n")
	)
	if i != -1 {
		long = long[:i]
	}
	long = strings.Join(strings.Fields(long), " ")
	if long == "" {
		return ""
	}
	rs := []rune(long)
	rs[0] = unicode.ToUpper(rs[0])
	if len(rs) > 80 {
		rs = append(rs[:77], []rune("...")...)
	} else if len(rs) > 1 && rs[len(rs)-1] == '.' {
		// Clip the period at the end by convention but only if the last but one
		// character is a letter or a digit. TODO: other cases?
		if r := rs[len(rs)-2]; unicode.IsLetter(r) || unicode.IsDigit(r) {
			rs = rs[:len(rs)-1]
		}
	}
	return string(rs)
}

func (md *Metadata) Usage(name string, args []ParamType) string {
	if md == nil {
		return strings.ToLower(name)
	}
	if usage, ok := md.raw.Directives["usage"]; ok {
		return usage
	}
	return strings.ToLower(name) + ParamsUsage(md.raw.Params, args)
}

func (md *Metadata) Child(name string) *Metadata {
	if md == nil {
		return nil
	}
	if md.children == nil {
		md.children = map[string]*Metadata{}
	}
	child, ok := md.children[name]
	if !ok {
		child = &Metadata{root: md.root, raw: md.raw.Child(name)}
		md.children[name] = child
	}
	return child
}
