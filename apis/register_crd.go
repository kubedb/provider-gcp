package apis

import (
	extentionapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		extentionapi.AddToScheme,
	)
}

// AddToScheme adds all Resources to the Scheme
func AddToSchemeCrd(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
