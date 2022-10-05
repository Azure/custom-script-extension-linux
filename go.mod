module github.com/Azure/custom-script-extension-linux

go 1.16

require (
	github.com/Azure/azure-extension-foundation v0.0.0-20190726000431-02f4f599e64a
	github.com/Azure/azure-extension-platform v0.0.0-20220913231409-868d66a2f29a
	github.com/Azure/azure-sdk-for-go v3.1.0-beta.0.20160802173609-87de771fcdf5+incompatible
	github.com/ahmetalpbalkan/go-httpbin v0.0.0-20160706084156-8817b883dae1
	github.com/go-kit/kit v0.12.0
	github.com/go-stack/stack v1.5.2 // indirect
	github.com/gorilla/context v0.0.0-20160525203319-aed02d124ae4 // indirect
	github.com/gorilla/mux v0.0.0-20160605233521-9fa818a44c2b // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20151027082146-e0fe6f683076 // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20150808065054-e02fc20de94c // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20160623135812-c539bca196be
	golang.org/x/text v0.3.7
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/go-kit/kit => github.com/go-kit/kit v0.1.1-0.20160721083846-b076b44dbec2