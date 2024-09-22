package entities

type Classifier struct {
	*ClassifierConfig
}

type ClassifierConfig struct {
	Type           string   `json:"type"`
	Conditions     []string `json:"conditions"`
	Classification string   `json:"classification"`
}

func NewClassifier(config *ClassifierConfig) *Classifier {
	return &Classifier{
		config,
	}
}
