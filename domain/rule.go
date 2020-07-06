package domain

type (
	Rule struct {
		ID        string
		Path      string
		Method    string
		Variable  Variable
		Weight    Weight
		Responses []byte
	}

	Variable map[string]interface{}

	WeightingFactor map[string]uint

	Weight map[string]WeightingFactor
)
