module github.com/rishi-anand/terraform-provider-kubernetes

go 1.15

require (
	github.com/hashicorp/terraform-plugin-docs v0.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.4.2
	github.com/hashicorp/yamux v0.0.0-20200609203250-aecfd211c9ce // indirect
	k8s.io/apimachinery v0.17.9
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200411171748-3d5a2fe318e4 // indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.17.9
