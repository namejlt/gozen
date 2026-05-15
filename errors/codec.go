package errors

import "gopkg.in/yaml.v3"

// CodeList is the YAML/JSON container for code definitions.
type CodeList struct {
	Codes map[string]string `yaml:"Codes" json:"codes"`
}

// LoadCodesFromYAML populates the registry from YAML bytes.
func LoadCodesFromYAML(data []byte) error {
	var cl CodeList
	if err := yaml.Unmarshal(data, &cl); err != nil {
		return err
	}
	for k, v := range cl.Codes {
		SetMessage(parseCode(k), v)
	}
	return nil
}

// parseCode wraps strconv.Atoi, returning 0 on parse failure.
func parseCode(s string) int {
	// local helper avoids importing strconv here (already used in errors.go).
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
