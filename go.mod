module github.com/Azure/custom-script-extension-linux

go 1.23

toolchain go1.23.4

require (
	github.com/Azure/azure-extension-foundation v0.0.0-20230404211847-9858bdd5c187
	github.com/Azure/azure-extension-platform v0.0.0-20241219234143-33858f5985a6
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/ahmetalpbalkan/go-httpbin v0.0.0-20160706084156-8817b883dae1
	github.com/go-kit/kit v0.13.0
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.2
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/text v0.21.0
)

require (
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.24 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.5.1 // indirect
	github.com/gorilla/context v0.0.0-20160525203319-aed02d124ae4 // indirect
	github.com/gorilla/mux v0.0.0-20160605233521-9fa818a44c2b // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/go-kit/kit => github.com/go-kit/kit v0.1.1-0.20160721083846-b076b44dbec2

// //DELETE after testing
// replace "github.com/Azure/custom-script-extension-linux" => "../custom-script-extension-linux"

