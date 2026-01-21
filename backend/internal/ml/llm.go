package ml

import "context"

type LLM interface {
	Generate(context.Context, HybridInput) (string, error)
}
