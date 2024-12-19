package rdf2go

import (
	"encoding/json"

	"github.com/nvkp/turtle"
)

func Unmarshal(data []byte, v interface{}) error {
	triples := []SimpleTriple{}

	err := turtle.Unmarshal(
		[]byte(data),
		&triples,
	)

	if err != nil {
		return err
	}

	result := convertToNestedMap(triples)

	// Serializa o resultado para JSON
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return err
	}

	// Deserializa o JSON resultante
	if err := json.Unmarshal(jsonResult, v); err != nil {
		return err
	}

	return nil
}

// convertToNestedMap converte uma lista de triplas em um mapa aninhado no formato desejado.
func convertToNestedMap(triples []SimpleTriple) map[string]interface{} {
	// Função recursiva para processar o mapa agrupado
	var processEntity func(subject string) map[string]interface{}

	// Função auxiliar para processar um objeto aninhado com chave "key"
	processNestedObject := func(subject string) map[string]interface{} {
		nested := processEntity(subject)
		nested["key"] = getKeyFromURI(subject)
		return nested
	}

	// Mapa para agrupar propriedades por sujeito
	grouped := make(map[string]map[string][]string)

	// Agrupando os atributos por sujeito
	for _, triple := range triples {
		if _, exists := grouped[triple.Subject]; !exists {
			grouped[triple.Subject] = make(map[string][]string)
		}
		grouped[triple.Subject][triple.Predicate] = append(grouped[triple.Subject][triple.Predicate], triple.Object)
	}

	processEntity = func(subject string) map[string]interface{} {
		result := make(map[string]interface{})
		for predicate, objects := range grouped[subject] {
			key := getKeyFromURI(predicate)

			if len(objects) == 1 {
				if subject == objects[0] {
					result[key] = objects[0]
				} else if grouped[objects[0]] != nil { // Verifica se é um objeto aninhado
					items := []map[string]interface{}{
						processNestedObject(objects[0]),
					}
					result[key] = items
				} else {
					result[key] = objects[0]
				}
			} else {
				var items []interface{}
				for _, obj := range objects {
					if grouped[obj] != nil { // Verifica se é um objeto aninhado
						items = append(items, processNestedObject(obj))
					} else {
						items = append(items, obj)
					}
				}
				result[key] = items
			}
		}

		return result
	}

	return processEntity(triples[0].Subject)
}

func getKeyFromURI(uri string) string {
	parts := split(uri, "/#")
	return parts[len(parts)-1]
}

func split(s string, delimiters string) []string {
	var result []string
	word := ""
	for _, char := range s {
		if containsRune(delimiters, char) {
			if word != "" {
				result = append(result, word)
				word = ""
			}
		} else {
			word += string(char)
		}
	}
	if word != "" {
		result = append(result, word)
	}
	return result
}

func containsRune(delimiters string, r rune) bool {
	for _, d := range delimiters {
		if d == r {
			return true
		}
	}
	return false
}
