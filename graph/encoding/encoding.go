package encoding

import (
	"encoding/json"

	"github.com/synful/cluster-simulate/graph"
)

func Unmarshal(data []byte) (graph.GraphDef, error) {
	def := graph.GraphDef{}
	err := json.Unmarshal(data, &def)
	if err != nil {
		return graph.GraphDef{}, err
	}
	return def, nil
}

func Marshal(def graph.GraphDef) ([]byte, error) {
	return json.Marshal(def)
}
