package rdf2go

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	rdf "github.com/deiu/gon3"
	jsonld "github.com/linkeddata/gojsonld"
)

// Graph structure
type Graph struct {
	triples    []*Triple
	httpClient *http.Client
	uri        string
	term       Term
	namespaces map[string]string
}

// NewHttpClient creates an http.Client to be used for parsing resources
// directly from the Web
func NewHttpClient(skip bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skip,
			},
		},
	}
}

// NewGraph creates a Graph object
func NewGraph(uri string, skipVerify ...bool) *Graph {
	skip := false
	if len(skipVerify) > 0 {
		skip = skipVerify[0]
	}
	g := &Graph{
		triples:    make([]*Triple, 0),
		httpClient: NewHttpClient(skip),
		uri:        uri,
		term:       NewResource(uri),
		namespaces: map[string]string{},
	}
	return g
}

// Len returns the length of the graph as number of triples in the graph
func (g *Graph) Len() int {
	return len(g.triples)
}

// Term returns a Graph Term object
func (g *Graph) Term() Term {
	return g.term
}

// URI returns a Graph URI object
func (g *Graph) URI() string {
	return g.uri
}

// One returns one triple based on a triple pattern of S, P, O objects
func (g *Graph) One(s Term, p Term, o Term) *Triple {
	for _, triple := range g.IterTriples() {
		if s != nil {
			if p != nil {
				if o != nil {
					if triple.Subject.Equal(s) && triple.Predicate.Equal(p) && triple.Object.Equal(o) {
						return triple
					}
				} else {
					if triple.Subject.Equal(s) && triple.Predicate.Equal(p) {
						return triple
					}
				}
			} else {
				if triple.Subject.Equal(s) {
					return triple
				}
			}
		} else if p != nil {
			if o != nil {
				if triple.Predicate.Equal(p) && triple.Object.Equal(o) {
					return triple
				}
			} else {
				if triple.Predicate.Equal(p) {
					return triple
				}
			}
		} else if o != nil {
			if triple.Object.Equal(o) {
				return triple
			}
		} else {
			return triple
		}
	}
	return nil
}

// IterTriples provides a channel containing all the triples in the graph.
// Note that the returned channel is already closed.
func (g *Graph) IterTriples() []*Triple {
	return g.triples
}

// Add is used to add a Triple object to the graph
func (g *Graph) Add(t *Triple) {
	if !slices.Contains(g.triples, t) {
		g.triples = append(g.triples, t)
	}
}

// AddTriple is used to add a triple made of individual S, P, O objects
func (g *Graph) AddTriple(s Term, p Term, o Term) {
	g.triples = append(g.triples, NewTriple(s, p, o))
}

// Remove is used to remove a Triple object
func (g *Graph) Remove(t *Triple) {
	for i, triple := range g.triples {
		if triple.Equal(t) {
			g.triples = append(g.triples[:i], g.triples[i+1:]...)
			break
		}
	}
}

// All is used to return all triples that match a given pattern of S, P, O objects
func (g *Graph) All(s Term, p Term, o Term) []*Triple {
	var triples []*Triple
	for _, triple := range g.IterTriples() {
		if s != nil {
			if p != nil {
				if o != nil {
					if triple.Subject.Equal(s) && triple.Predicate.Equal(p) && triple.Object.Equal(o) {
						triples = append(triples, triple)
					}
				} else {
					if triple.Subject.Equal(s) && triple.Predicate.Equal(p) {
						triples = append(triples, triple)
					}
				}
			} else {
				if triple.Subject.Equal(s) {
					triples = append(triples, triple)
				}
			}
		} else if p != nil {
			if o != nil {
				if triple.Predicate.Equal(p) && triple.Object.Equal(o) {
					triples = append(triples, triple)
				}
			} else {
				if triple.Predicate.Equal(p) {
					triples = append(triples, triple)
				}
			}
		} else if o != nil {
			if triple.Object.Equal(o) {
				triples = append(triples, triple)
			}
		}
	}
	return triples
}

// Merge is used to add all the triples form another graph to this one
func (g *Graph) Merge(toMerge *Graph) {
	for _, triple := range toMerge.IterTriples() {
		g.Add(triple)
	}
}

// Parse is used to parse RDF data from a reader, using the provided mime type
func (g *Graph) Parse(reader io.Reader, mime string) error {
	parserName := mimeParser[mime]
	if len(parserName) == 0 {
		parserName = "guess"
	}
	if parserName == "jsonld" {
		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
		jsonData, err := jsonld.ReadJSON(buf.Bytes())
		if err != nil {
			return err
		}
		options := &jsonld.Options{}
		options.Base = ""
		options.ProduceGeneralizedRdf = false
		dataSet, err := jsonld.ToRDF(jsonData, options)
		if err != nil {
			return err
		}
		for t := range dataSet.IterTriples() {
			g.AddTriple(jterm2term(t.Subject), jterm2term(t.Predicate), jterm2term(t.Object))
		}

	} else if parserName == "turtle" {
		parser, err := rdf.NewParser(g.uri).Parse(reader)
		if err != nil {
			return err
		}
		for s := range parser.IterTriples() {
			g.AddTriple(rdf2term(s.Subject), rdf2term(s.Predicate), rdf2term(s.Object))
		}
	} else {
		return errors.New(parserName + " is not supported by the parser")
	}
	return nil
}

// LoadURI is used to load RDF data from a specific URI
func (g *Graph) LoadURI(uri string) error {
	doc := defrag(uri)
	q, err := http.NewRequest("GET", doc, nil)
	if err != nil {
		return err
	}
	if len(g.uri) == 0 {
		g.uri = doc
	}
	q.Header.Set("Accept", "text/turtle;q=1,application/ld+json;q=0.5")
	r, err := g.httpClient.Do(q)
	if err != nil {
		return err
	}
	if r != nil {
		defer r.Body.Close()
		if r.StatusCode == 200 {
			g.Parse(r.Body, r.Header.Get("Content-Type"))
		} else {
			return fmt.Errorf("Could not fetch graph from %s - HTTP %d", uri, r.StatusCode)
		}
	}
	return nil
}

// String is used to serialize the graph object using NTriples
func (g *Graph) String() string {
	var toString string

	// add namespaces
	for ns := range g.namespaces {
		toString += fmt.Sprintf("@prefix %s <%s> .\n", ns, g.namespaces[ns])
	}

	if len(g.namespaces) > 0 {
		toString += "\n"
	}

	for _, triple := range g.IterTriples() {
		toString += triple.String() + "\n"
	}
	return toString
}

// Serialize is used to serialize a graph based on a given mime type
func (g *Graph) Serialize(w io.Writer, mime string) error {
	serializerName := mimeSerializer[mime]
	if serializerName == "jsonld" {
		return g.serializeJSONLD(w)
	}
	// just return Turtle by default
	return g.serializeTurtle(w)
}

// @TODO improve streaming
func (g *Graph) serializeTurtle(w io.Writer) error {
	var err error

	for ns := range g.namespaces {
		_, err = fmt.Fprintf(w, "@prefix %s <%s> .\n", ns, g.namespaces[ns])
		if err != nil {
			return err
		}
	}

	if len(g.namespaces) > 0 {
		_, err = fmt.Fprintf(w, "\n")
		if err != nil {
			return err
		}
	}

	triplesBySubject := make(map[string][]*Triple)
	subjects := []string{}

	for _, triple := range g.IterTriples() {
		s := encodeTerm(triple.Subject)
		triplesBySubject[s] = append(triplesBySubject[s], triple)

		if !slices.Contains(subjects, s) {
			subjects = append(subjects, s)
		}
	}

	for i, subject := range subjects {
		_, err = fmt.Fprintf(w, "%s\n", subject)
		if err != nil {
			return err
		}

		triples := triplesBySubject[subject]

		for key, triple := range triples {
			p := encodeTerm(triple.Predicate)
			o := encodeTerm(triple.Object)

			if key == len(triples)-1 {
				_, err = fmt.Fprintf(w, "  %s %s .", p, o)
				if err != nil {
					return err
				}
				break
			}

			_, err = fmt.Fprintf(w, "  %s %s ;\n", p, o)
			if err != nil {
				return err
			}
		}

		if len(subjects) > i+1 {
			_, err = fmt.Fprintf(w, "\n\n")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Graph) serializeJSONLD(w io.Writer) error {
	r := []map[string]interface{}{}
	for _, elt := range g.IterTriples() {
		var one map[string]interface{}
		switch elt.Subject.(type) {
		case *BlankNode:
			one = map[string]interface{}{
				"@id": elt.Subject.(*BlankNode).String(),
			}
		default:
			one = map[string]interface{}{
				"@id": elt.Subject.(*Resource).URI,
			}
		}
		switch t := elt.Object.(type) {
		case *Resource:
			one[elt.Predicate.(*Resource).URI] = []map[string]string{
				{
					"@id": t.URI,
				},
			}
			break
		case *Literal:
			v := map[string]string{
				"@value": t.Value,
			}
			if t.Datatype != nil && len(t.Datatype.String()) > 0 {
				v["@type"] = debrack(t.Datatype.String())
			}
			if len(t.Language) > 0 {
				v["@language"] = t.Language
			}
			one[elt.Predicate.(*Resource).URI] = []map[string]string{v}
		}
		r = append(r, one)
	}
	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, string(bytes))
	return nil
}

func (g *Graph) Bind(ns *Namespace) {
	namespace := ns.NS

	if !strings.HasSuffix(namespace, ":") {
		namespace += ":"
	}

	g.namespaces[namespace] = ns.URI
}
