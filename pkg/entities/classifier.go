package entities

type Classifier struct {
	*ClassifierConfig
}

type ClassifierConfig struct {
	Type string `json:"type"`
	CEL  string `json:"cel"`
}
