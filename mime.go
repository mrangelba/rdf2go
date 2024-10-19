package rdf2go

import (
	"regexp"
)

const (
	TurtleMime = "text/turtle"
	JsonldMime = "application/ld+json"
	SparQLMime = "application/sparql-update"
	N3Mime     = "text/n3"
	RDFXMLMime = "application/rdf+xml"
)

var mimeParser = map[string]string{
	TurtleMime: "turtle",
	JsonldMime: "jsonld",
	SparQLMime: "internal",
}

var mimeSerializer = map[string]string{
	"application/ld+json": "jsonld",
	"text/html":           "internal",
}

var RdfExtMime = map[string]string{
	".ttl":    TurtleMime,
	".n3":     N3Mime,
	".rdf":    RDFXMLMime,
	".jsonld": JsonldMime,
}

var rdfExtensions = []string{
	".ttl",
	".n3",
	".rdf",
	".jsonld",
}

var (
	serializerMimes = []string{}
	validMimeType   = regexp.MustCompile(`^\w+/\w+$`)
)
