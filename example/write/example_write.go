package main

import (
	"bytes"
	"fmt"
	"os"

	rdf "github.com/mrangelba/rdf2go"
)

func main() {
	baseUri := "http://schema.org/"
	ns1 := rdf.NewNamespace("ns1", "http://schema.org/")

	g := rdf.NewGraph(baseUri)
	g.Bind(ns1)

	profile := rdf.NewResource("http://solid/profile/card#me")
	credential1 := rdf.NewResource("#credential1")
	credential2 := rdf.NewResource("#credential2")
	skill1 := rdf.NewResource("#skill1")
	skill2 := rdf.NewResource("#skill2")
	skill3 := rdf.NewResource("#skill3")

	g.AddTriple(profile, rdf.NewRDFType(), ns1.WithAttr("Profile"))
	g.AddTriple(profile, ns1.WithAttr("description"), rdf.NewLiteral("ShortBiography"))
	g.AddTriple(profile, ns1.WithAttr("credentials"), rdf.NewList(credential1, credential2))
	g.AddTriple(profile, ns1.WithAttr("id"), rdf.NewLiteral("01"))
	g.AddTriple(profile, ns1.WithAttr("skills"), rdf.NewList(skill1, skill2, skill3))
	g.AddTriple(profile, ns1.WithAttr("name"), rdf.NewLiteral("John Doe"))

	g.AddTriple(credential1, rdf.NewRDFType(), ns1.WithAttr("Credential"))
	g.AddTriple(credential1, ns1.WithAttr("issuedBy"), rdf.NewLiteral("CredentialIssuer"))
	g.AddTriple(credential1, ns1.WithAttr("description"), rdf.NewLiteral("Description"))

	g.AddTriple(credential2, rdf.NewRDFType(), ns1.WithAttr("Credential"))
	g.AddTriple(credential2, ns1.WithAttr("issuedBy"), rdf.NewLiteral("CredentialIssuer"))
	g.AddTriple(credential2, ns1.WithAttr("description"), rdf.NewLiteral("Description"))

	g.AddTriple(skill1, rdf.NewRDFType(), ns1.WithAttr("Skill"))
	g.AddTriple(skill1, ns1.WithAttr("name"), rdf.NewLiteral("Leadership"))
	g.AddTriple(skill1, ns1.WithAttr("alternateName"), rdf.NewLiteral("Skill 1"))

	g.AddTriple(skill2, rdf.NewRDFType(), ns1.WithAttr("Skill"))
	g.AddTriple(skill2, ns1.WithAttr("name"), rdf.NewLiteral("Teamwork"))
	g.AddTriple(skill2, ns1.WithAttr("alternateName"), rdf.NewLiteral("Skill 2"))

	g.AddTriple(skill3, rdf.NewRDFType(), ns1.WithAttr("Skill"))
	g.AddTriple(skill3, ns1.WithAttr("name"), rdf.NewLiteral("Communication"))
	g.AddTriple(skill3, ns1.WithAttr("alternateName"), rdf.NewLiteral("Skill 3"))

	buf := new(bytes.Buffer)

	err := g.Serialize(buf, "text/turtle")

	if err != nil {
		panic(err)
	}

	fmt.Println(buf.String())

	// Save to file:
	f, err := os.Create("example.ttl")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	err = g.Serialize(f, rdf.TurtleMime)
	if err != nil {
		panic(err)
	}
}
