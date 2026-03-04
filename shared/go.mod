module github.com/code-together/shared

go 1.25

require (
	github.com/oapi-codegen/runtime v1.1.2
	github.com/tidwall/gjson v1.18.0
	github.com/tidwall/sjson v1.2.5
)

// Tool modules
replace github.com/code-together/shared/tools/hook-browser => ./tools/hook-browser

replace github.com/code-together/shared/tools/hook-collector => ./tools/hook-collector

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
)
