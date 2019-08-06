package definitions

// ResourcesFactory gather all resources definitions from this package to make it ease for clients to use this Factory as a single point of reference
type ResourcesFactory struct {
	RoleBinding      *roleBindingResource
	BuildConfig      *buildConfigResource
	DeploymentConfig *deploymentConfigResource
	Route            *routeResource
}
