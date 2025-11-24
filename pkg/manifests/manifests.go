package manifests

import (
	"fmt"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1 "github.com/openshift/api/security/v1"

	"github.com/Tal-or/dra-deployer/pkg/assets"
)

var (
	manifestScheme = runtime.NewScheme()
	codecFactory   serializer.CodecFactory
	runtimeDecoder runtime.Decoder
)

func init() {
	utilruntime.Must(corev1.AddToScheme(manifestScheme))
	utilruntime.Must(rbacv1.AddToScheme(manifestScheme))
	utilruntime.Must(appsv1.AddToScheme(manifestScheme))
	utilruntime.Must(admissionregistrationv1.AddToScheme(manifestScheme))
	utilruntime.Must(resourcev1beta1.AddToScheme(manifestScheme))
	utilruntime.Must(securityv1.Install(manifestScheme))
	codecFactory = serializer.NewCodecFactory(manifestScheme)
	runtimeDecoder = codecFactory.UniversalDeserializer()
}

// Manifests holds all DRA CPU driver manifest objects
type Manifests struct {
	ServiceAccount                   *corev1.ServiceAccount
	SecurityContextConstraints       *securityv1.SecurityContextConstraints
	ClusterRole                      *rbacv1.ClusterRole
	ClusterRoleBinding               *rbacv1.ClusterRoleBinding
	DaemonSet                        *appsv1.DaemonSet
	DeviceClasses                    []*resourcev1beta1.DeviceClass
	ValidatingAdmissionPolicy        *admissionregistrationv1.ValidatingAdmissionPolicy
	ValidatingAdmissionPolicyBinding *admissionregistrationv1.ValidatingAdmissionPolicyBinding
}

func (m *Manifests) GetObjects() []client.Object {
	return []client.Object{
		m.ServiceAccount,
		m.SecurityContextConstraints,
		m.ClusterRole,
		m.ClusterRoleBinding,
		m.DaemonSet,
		m.ValidatingAdmissionPolicy,
		m.ValidatingAdmissionPolicyBinding,
	}
}

// GetAll returns all DRA CPU driver manifests with the specified namespace and image applied
// where applicable. Cluster-scoped resources (ClusterRole, DeviceClasses, etc.) are
// not affected by the namespace parameter.
func GetAll(namespace, image string) (*Manifests, error) {
	sa, err := GetServiceAccount(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get ServiceAccount: %w", err)
	}

	cr, err := GetClusterRole()
	if err != nil {
		return nil, fmt.Errorf("failed to get ClusterRole: %w", err)
	}

	crb, err := GetClusterRoleBinding(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get ClusterRoleBinding: %w", err)
	}

	ds, err := GetDaemonSet(namespace, image)
	if err != nil {
		return nil, fmt.Errorf("failed to get DaemonSet: %w", err)
	}

	vap, err := GetValidatingAdmissionPolicy(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get ValidatingAdmissionPolicy: %w", err)
	}

	vapb, err := GetValidatingAdmissionPolicyBinding()
	if err != nil {
		return nil, fmt.Errorf("failed to get ValidatingAdmissionPolicyBinding: %w", err)
	}

	scc, err := GetSecurityContextConstraints(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get SecurityContextConstraints: %w", err)
	}

	return &Manifests{
		ServiceAccount:                   sa,
		SecurityContextConstraints:       scc,
		ClusterRole:                      cr,
		ClusterRoleBinding:               crb,
		DaemonSet:                        ds,
		ValidatingAdmissionPolicy:        vap,
		ValidatingAdmissionPolicyBinding: vapb,
	}, nil
}

// decodeManifest reads and decodes a YAML manifest file into a Kubernetes object
func decodeManifest(filename string, obj runtime.Object) error {
	data, err := assets.Yamls.ReadFile(fmt.Sprintf("yamls/%s", filename))
	if err != nil {
		return fmt.Errorf("failed to read manifest %s: %w", filename, err)
	}

	_, _, err = runtimeDecoder.Decode(data, nil, obj)
	if err != nil {
		return fmt.Errorf("failed to decode manifest %s: %w", filename, err)
	}

	return nil
}

// GetServiceAccount returns the DRA CPU driver ServiceAccount
func GetServiceAccount(namespace string) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	if err := decodeManifest("serviceaccount.yaml", sa); err != nil {
		return nil, err
	}
	sa.Namespace = namespace
	return sa, nil
}

// GetClusterRole returns the DRA CPU driver ClusterRole
func GetClusterRole() (*rbacv1.ClusterRole, error) {
	cr := &rbacv1.ClusterRole{}
	if err := decodeManifest("clusterrole.yaml", cr); err != nil {
		return nil, err
	}
	return cr, nil
}

// GetClusterRoleBinding returns the DRA CPU driver ClusterRoleBinding
func GetClusterRoleBinding(namespace string) (*rbacv1.ClusterRoleBinding, error) {
	crb := &rbacv1.ClusterRoleBinding{}
	if err := decodeManifest("clusterrolebinding.yaml", crb); err != nil {
		return nil, err
	}
	crb.Subjects[0].Namespace = namespace
	return crb, nil
}

// GetDaemonSet returns the DRA CPU driver DaemonSet
func GetDaemonSet(namespace, image string) (*appsv1.DaemonSet, error) {
	ds := &appsv1.DaemonSet{}
	if err := decodeManifest("daemonset.yaml", ds); err != nil {
		return nil, err
	}
	ds.Namespace = namespace
	// Update the container image
	if len(ds.Spec.Template.Spec.Containers) > 0 {
		ds.Spec.Template.Spec.Containers[0].Image = image
		klog.V(4).InfoS("Set DaemonSet container image", "image", image)
	}
	return ds, nil
}

// GetDeviceClasses returns the DRA CPU driver DeviceClasses (both exclusive-cpu and shared-cpu)
// Note: This file contains multiple DeviceClass objects separated by ---
// Returns a slice of DeviceClass objects
func GetDeviceClasses() ([]*resourcev1beta1.DeviceClass, error) {
	data, err := assets.Yamls.ReadFile("yamls/deviceclass.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read deviceclass.yaml: %w", err)
	}

	// Split the YAML document by --- separator
	decoder := serializer.NewCodecFactory(manifestScheme).UniversalDeserializer()

	var deviceClasses []*resourcev1beta1.DeviceClass

	// Decode multiple documents
	// Simple approach: split by "---" and decode each
	docs := splitYAMLDocuments(data)
	for _, doc := range docs {
		if len(doc) == 0 {
			continue
		}
		dc := &resourcev1beta1.DeviceClass{}
		_, _, err := decoder.Decode(doc, nil, dc)
		if err != nil {
			return nil, fmt.Errorf("failed to decode DeviceClass: %w", err)
		}
		deviceClasses = append(deviceClasses, dc)
	}

	return deviceClasses, nil
}

// GetExclusiveCPUDeviceClass returns the exclusive-cpu DeviceClass
func GetExclusiveCPUDeviceClass() (*resourcev1beta1.DeviceClass, error) {
	deviceClasses, err := GetDeviceClasses()
	if err != nil {
		return nil, err
	}

	for _, dc := range deviceClasses {
		if dc.Name == "exclusive-cpu" {
			return dc, nil
		}
	}

	return nil, fmt.Errorf("exclusive-cpu DeviceClass not found")
}

// GetSharedCPUDeviceClass returns the shared-cpu DeviceClass
func GetSharedCPUDeviceClass() (*resourcev1beta1.DeviceClass, error) {
	deviceClasses, err := GetDeviceClasses()
	if err != nil {
		return nil, err
	}

	for _, dc := range deviceClasses {
		if dc.Name == "shared-cpu" {
			return dc, nil
		}
	}

	return nil, fmt.Errorf("shared-cpu DeviceClass not found")
}

// GetValidatingAdmissionPolicy returns the DRA CPU driver ValidatingAdmissionPolicy
func GetValidatingAdmissionPolicy(namespace string) (*admissionregistrationv1.ValidatingAdmissionPolicy, error) {
	vap := &admissionregistrationv1.ValidatingAdmissionPolicy{}
	if err := decodeManifest("validatingadmissionpolicy.yaml", vap); err != nil {
		return nil, err
	}

	// Add the matchCondition with the correct namespace
	// The expression references: system:serviceaccount:<namespace>:<service-account-name>
	matchCondition := admissionregistrationv1.MatchCondition{
		Name:       "isRestrictedUser",
		Expression: fmt.Sprintf(`request.userInfo.username == "system:serviceaccount:%s:dra-cpu-driver-service-account"`, namespace),
	}
	vap.Spec.MatchConditions = append(vap.Spec.MatchConditions, matchCondition)

	klog.V(4).InfoS("Added ValidatingAdmissionPolicy matchCondition",
		"name", matchCondition.Name,
		"expression", matchCondition.Expression)

	return vap, nil
}

// GetValidatingAdmissionPolicyBinding returns the DRA CPU driver ValidatingAdmissionPolicyBinding
func GetValidatingAdmissionPolicyBinding() (*admissionregistrationv1.ValidatingAdmissionPolicyBinding, error) {
	vapb := &admissionregistrationv1.ValidatingAdmissionPolicyBinding{}
	if err := decodeManifest("validatingadmissionpolicybinding.yaml", vapb); err != nil {
		return nil, err
	}
	return vapb, nil
}

// GetSecurityContextConstraints returns the DRA CPU driver SecurityContextConstraints
func GetSecurityContextConstraints(namespace string) (*securityv1.SecurityContextConstraints, error) {
	scc := &securityv1.SecurityContextConstraints{}
	if err := decodeManifest("securitycontextconstraints.yaml", scc); err != nil {
		return nil, err
	}
	scc.Users = []string{fmt.Sprintf("system:serviceaccount:%s:dra-cpu-driver-service-account", namespace)}
	return scc, nil
}

// splitYAMLDocuments splits a YAML file with multiple documents (separated by ---)
// into individual documents
func splitYAMLDocuments(data []byte) [][]byte {
	var documents [][]byte
	separator := []byte("\n---\n")

	docs := [][]byte{}
	current := []byte{}

	lines := []byte{}
	for i, b := range data {
		lines = append(lines, b)
		if b == '\n' || i == len(data)-1 {
			if len(lines) >= 5 && string(lines[len(lines)-5:len(lines)-1]) == "\n---" {
				// Found separator
				if len(current) > 0 {
					docs = append(docs, current)
				}
				current = []byte{}
				lines = []byte{}
			} else {
				current = append(current, lines...)
				lines = []byte{}
			}
		}
	}

	if len(current) > 0 {
		docs = append(docs, current)
	}

	// Simpler alternative using bytes split
	parts := [][]byte{data}
	if len(separator) > 0 {
		parts = splitBytes(data, separator)
	}

	for _, part := range parts {
		trimmed := trimSpace(part)
		if len(trimmed) > 0 {
			documents = append(documents, trimmed)
		}
	}

	return documents
}

// splitBytes splits a byte slice by a separator
func splitBytes(data, separator []byte) [][]byte {
	if len(separator) == 0 {
		return [][]byte{data}
	}

	var parts [][]byte
	for len(data) > 0 {
		idx := indexBytes(data, separator)
		if idx < 0 {
			parts = append(parts, data)
			break
		}
		parts = append(parts, data[:idx])
		data = data[idx+len(separator):]
	}
	return parts
}

// indexBytes finds the first occurrence of sep in data
func indexBytes(data, sep []byte) int {
	if len(sep) == 0 {
		return 0
	}
	for i := 0; i <= len(data)-len(sep); i++ {
		if matchBytes(data[i:], sep) {
			return i
		}
	}
	return -1
}

// matchBytes checks if the beginning of data matches pattern
func matchBytes(data, pattern []byte) bool {
	if len(data) < len(pattern) {
		return false
	}
	for i := 0; i < len(pattern); i++ {
		if data[i] != pattern[i] {
			return false
		}
	}
	return true
}

// trimSpace removes leading and trailing whitespace from a byte slice
func trimSpace(data []byte) []byte {
	start := 0
	end := len(data)

	for start < end && isSpace(data[start]) {
		start++
	}

	for start < end && isSpace(data[end-1]) {
		end--
	}

	return data[start:end]
}

// isSpace checks if a byte is a whitespace character
func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
