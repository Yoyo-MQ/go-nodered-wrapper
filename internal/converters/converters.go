package converters

import (
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
)

// ConvertFlowToNodeRedFormat converts a flow definition to Node-RED format
func ConvertFlowToNodeRedFormat(flow *types.FlowDefinition) []map[string]interface{} {
	var nodeRedNodes []map[string]interface{}

	tabNode := map[string]interface{}{
		"id":    flow.ID,
		"type":  "tab",
		"label": flow.Name,
	}
	if flow.Description != "" {
		tabNode["info"] = flow.Description
	}
	nodeRedNodes = append(nodeRedNodes, tabNode)

	for _, node := range flow.Nodes {
		nodeRedNode := map[string]interface{}{
			"id":    node.ID,
			"type":  node.Type,
			"name":  node.Name,
			"x":     node.Position.X,
			"y":     node.Position.Y,
			"z":     flow.ID,
			"wires": node.Wires,
		}

		for key, value := range node.Properties {
			nodeRedNode[key] = value
		}

		nodeRedNodes = append(nodeRedNodes, nodeRedNode)
	}

	return nodeRedNodes
}
