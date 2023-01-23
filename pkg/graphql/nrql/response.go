package nrql

// --- GraphQL for NRQL query --- //
type GraphQlNrqlResponse[T interface{}] struct {
	Data   Data[T]     `json:"data"`
	Errors interface{} `json:"errors"`
}

type Data[T interface{}] struct {
	Actor Actor[T] `json:"actor"`
}

type Actor[T interface{}] struct {
	Nrql Nrql[T] `json:"nrql"`
}

type Nrql[T interface{}] struct {
	Results []T `json:"results"`
}
