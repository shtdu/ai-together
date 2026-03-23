module github.com/code-together/bdd

go 1.25.0

// Local module replacements for development
replace github.com/code-together/shared => ../shared

replace github.com/code-together/integration => ../shared/integration

replace github.com/code-together/integration_manager => ../shared/integration_manager

require (
	github.com/code-together/integration_manager v0.0.0-00010101000000-000000000000
	github.com/code-together/shared v0.0.0-00010101000000-000000000000
	github.com/cucumber/godog v0.14.1
	github.com/google/uuid v1.6.0
	github.com/oapi-codegen/runtime v1.2.0
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/cucumber/gherkin/go/v26 v26.2.0 // indirect
	github.com/cucumber/messages/go/v21 v21.0.1 // indirect
	github.com/getkin/kin-openapi v0.131.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gofrs/uuid v4.3.1+incompatible // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
