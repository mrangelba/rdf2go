package rdf2go

import "strings"

type Namespace struct {
	NS  string
	URI string
}

func NewNamespace(ns string, uri string) *Namespace {
	if !strings.HasSuffix(ns, ":") {
		ns += ":"
	}

	return &Namespace{
		NS:  ns,
		URI: uri,
	}
}

func (ns *Namespace) WithAttr(name string) (term Term) {
	return Term(&NamespaceAttr{NS: ns.NS, Attr: name})
}
