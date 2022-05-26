//go:build viper_yaml2
// +build viper_yaml2

package yaml

// original form of the data
const original = `# key-value pair
key: value
list:
- item1
- item2
- item3
map:
  key: value

# nested
# map
nested_map:
  map:
    key: value
    list:
    - item1
    - item2
    - item3
`

// encoded form of the data
const encoded = `key: value
list:
- item1
- item2
- item3
map:
  key: value
nested_map:
  map:
    key: value
    list:
    - item1
    - item2
    - item3
`

// decoded form of the data
//
// in case of YAML it's slightly different from Viper's internal representation
// (eg. map is decoded into a map with interface key)
var decoded = map[string]interface{}{
	"key": "value",
	"list": []interface{}{
		"item1",
		"item2",
		"item3",
	},
	"map": map[interface{}]interface{}{
		"key": "value",
	},
	"nested_map": map[interface{}]interface{}{
		"map": map[interface{}]interface{}{
			"key": "value",
			"list": []interface{}{
				"item1",
				"item2",
				"item3",
			},
		},
	},
}

// Viper's internal representation
var data = map[string]interface{}{
	"key": "value",
	"list": []interface{}{
		"item1",
		"item2",
		"item3",
	},
	"map": map[string]interface{}{
		"key": "value",
	},
	"nested_map": map[string]interface{}{
		"map": map[string]interface{}{
			"key": "value",
			"list": []interface{}{
				"item1",
				"item2",
				"item3",
			},
		},
	},
}
