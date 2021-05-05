module ganni-tool

go 1.15

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/jonboulle/clockwork v0.1.0 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/cli-runtime v0.21.0 // indirect
	k8s.io/client-go v0.21.0
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.21.0 // indirect

)

replace (
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
	k8s.io/kubectl => k8s.io/kubectl v0.19.0
)
