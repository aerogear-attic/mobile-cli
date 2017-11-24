
build:
	go build -o mobile ./cmd/mobile

generate:
	vendor/k8s.io/code-generator/generate-internal-groups.sh client github.com/feedhenry/mobile-cli/pkg/client/mobile github.com/feedhenry/mobile-cli/pkg/apis github.com/feedhenry/mobile-cli/pkg/apis  "mobile:v1alpha1"
	vendor/k8s.io/code-generator/generate-internal-groups.sh client github.com/feedhenry/mobile-cli/pkg/client/servicecatalog github.com/feedhenry/mobile-cli/pkg/apis github.com/feedhenry/mobile-cli/pkg/apis  "servicecatalog:v1beta1"