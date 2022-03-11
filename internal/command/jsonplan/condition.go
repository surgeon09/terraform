package jsonplan

type conditionResult struct {
	// Address is the absolute address of the condition's containing object
	Address      string `json:"address,omitempty"`
	Type         string `json:"condition_type,omitempty"`
	Result       bool   `json:"result"`
	Unknown      bool   `json:"unknown"`
	ErrorMessage string `json:"error_message,omitempty"`
}
