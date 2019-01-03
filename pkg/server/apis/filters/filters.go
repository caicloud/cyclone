package filters

import "github.com/caicloud/nirvana/service"

// Filters returns a list of filters.
func Filters() []service.Filter {
	return []service.Filter{
		service.RedirectTrailingSlash(),
		service.FillLeadingSlash(),
		service.ParseRequestForm(),
	}
}
