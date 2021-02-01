package kubernetes

import (
	"encoding/json"
	"io"
	"strings"

	k8s "github.com/rishi-anand/terraform-provider-kubernetes/client"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

type ManifestAction string

const (
	ApplyAction  ManifestAction = "apply"
	DeleteAction ManifestAction = "delete"
)

func doManifestAction(action ManifestAction, content, nsoverride string, skipResources []string) error {
	//Get Kube client
	kubeClient, err := k8s.GetClient()
	if err != nil {
		return err
	}

	_, apiCollections, err := kubeClient.Discovery().ServerGroupsAndResources()
	if err != nil {
		return err
	}

	clusterScopedResourceKindMap := make(map[string]bool)
	for _, apiCollection := range apiCollections {
		for _, apiResource := range apiCollection.APIResources {
			if !apiResource.Namespaced {
				clusterScopedResourceKindMap[apiResource.Kind] = true
			}
		}
	}

	//Discover Server Resource Types
	dd := kubeClient.Discovery()
	apigroups, err := restmapper.GetAPIGroupResources(dd)
	if err != nil {
		return err
	}

	//Create Map of resoruces
	rm := restmapper.NewDiscoveryRESTMapper(apigroups)

	//Decoder for the Yaml/Json file
	d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(content), 4096)

	//Read & Process Manifest Blocls from Manifest file one by one till EOF
	for {

		ext := runtime.RawExtension{}
		if err := d.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		//Gt Group Version Kid for the Manifest Block
		_, gvk, _ := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, nil)
		mapping, err := rm.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		restconfig, err := k8s.GetConfig()
		if err != nil {
			return err
		}

		//Get Dynamic Client for Specific Group, Version
		restconfig.GroupVersion = &schema.GroupVersion{
			Group:   mapping.GroupVersionKind.Group,
			Version: mapping.GroupVersionKind.Version,
		}
		dclient, err := k8s.GetDynamicClientForConfig(restconfig)
		if err != nil {
			return err
		}

		//Construct Unstructured Object to apply
		var unstruct unstructured.Unstructured
		unstruct.Object = make(map[string]interface{})
		var blob interface{}
		if err := json.Unmarshal(ext.Raw, &blob); err != nil {
			return err
		}

		//Handle Namespace and Object Name
		unstruct.Object = blob.(map[string]interface{})
		ns := "default"
		if nsoverride != "" {
			ns = nsoverride
		}

		//If Object had specified namespace, the override if override specified
		if md, ok := unstruct.Object["metadata"]; ok {
			metadata := md.(map[string]interface{})
			if internalns, ok := metadata["namespace"]; ok {
				objns := internalns.(string)
				if objns != "" && nsoverride != "" {
					//Override the namespace provided in the object
					metadata["namespace"] = ns
				}
			}
		}

		var res dynamic.ResourceInterface
		var skipResource = false
		for _, v := range skipResources {
			if v == mapping.GroupVersionKind.Kind {
				skipResource = true
				break
			}
		}

		if skipResource {
			continue
		}

		if clusterScopedResourceKindMap[mapping.GroupVersionKind.Kind] {
			res = dclient.Resource(mapping.Resource)
		} else {
			res = dclient.Resource(mapping.Resource).Namespace(ns)
		}

		getObj := metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       unstruct.GetKind(),
				APIVersion: unstruct.GetAPIVersion(),
			},
		}

		if action == ApplyAction {
			if unstructGet, err := res.Get(unstruct.GetName(), getObj); err != nil {
				if apierrs.IsNotFound(err) {
					if _, err = res.Create(&unstruct, metav1.CreateOptions{}); err != nil {
						return err
					}
				} else {
					return err
				}
			} else {

				unstruct.SetResourceVersion(unstructGet.GetResourceVersion())
				out, err := unstruct.MarshalJSON()
				if err != nil {
					return err
				}
				if _, err = res.Patch(unstruct.GetName(), types.MergePatchType, out, metav1.PatchOptions{}); err != nil {
					return err
				}
			}
		} else if action == DeleteAction {
			if err := res.Delete(unstruct.GetName(), &metav1.DeleteOptions{}); err != nil {
				if !apierrs.IsNotFound(err) {
					return err
				}
			}
		}

	}
	return nil
}
