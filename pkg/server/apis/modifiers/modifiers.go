// +nirvana:api=modifiers:"Modifiers"

package modifiers

import "github.com/caicloud/nirvana/service"

// Modifiers returns a list of modifiers.
func Modifiers() []service.DefinitionModifier {
	return []service.DefinitionModifier{
		service.FirstContextParameter(),
		service.ConsumeAllIfConsumesIsEmpty(),
		service.ProduceAllIfProducesIsEmpty(),
		service.ConsumeNoneForHTTPGet(),
		service.ConsumeNoneForHTTPDelete(),
		service.ProduceNoneForHTTPDelete(),
	}
}
