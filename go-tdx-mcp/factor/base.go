package factor

import "github.com/tdx/go-tdx-mcp/indicator"

type Factor interface {
	Name() string
	Category() string
	Description() string
	Inputs() []string
	Compute(bars []indicator.Bar) []float64
}

type FactorMeta struct {
	name        string
	category    string
	description string
	inputs      []string
	computeFn   func(bars []indicator.Bar) []float64
}

func (f *FactorMeta) Name() string           { return f.name }
func (f *FactorMeta) Category() string       { return f.category }
func (f *FactorMeta) Description() string    { return f.description }
func (f *FactorMeta) Inputs() []string       { return f.inputs }
func (f *FactorMeta) Compute(bars []indicator.Bar) []float64 { return f.computeFn(bars) }

var registry = make(map[string]*FactorMeta)

func Register(name, category, description string, inputs []string, fn func(bars []indicator.Bar) []float64) {
	if _, exists := registry[name]; exists {
		panic("factor already registered: " + name)
	}
	registry[name] = &FactorMeta{
		name:        name,
		category:    category,
		description: description,
		inputs:      inputs,
		computeFn:   fn,
	}
}

func Get(name string) *FactorMeta {
	return registry[name]
}

func List() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	return names
}

func ListByCategory(cat string) []string {
	names := make([]string, 0)
	for _, f := range registry {
		if f.category == cat {
			names = append(names, f.name)
		}
	}
	return names
}
