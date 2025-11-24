package manifests

import (
	"testing"
)

func TestGetAll(t *testing.T) {
	namespace := "test-namespace"
	image := "quay.io/test/test-image:v1.0.0"
	manifests, err := GetAll(namespace, image)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if manifests == nil {
		t.Fatal("GetAll returned nil")
	}

	// Verify ServiceAccount
	if manifests.ServiceAccount == nil {
		t.Error("ServiceAccount is nil")
	} else if manifests.ServiceAccount.Namespace != namespace {
		t.Errorf("ServiceAccount namespace: expected %s, got %s", namespace, manifests.ServiceAccount.Namespace)
	}

	// Verify ClusterRole
	if manifests.ClusterRole == nil {
		t.Error("ClusterRole is nil")
	}

	// Verify ClusterRoleBinding
	if manifests.ClusterRoleBinding == nil {
		t.Error("ClusterRoleBinding is nil")
	} else if len(manifests.ClusterRoleBinding.Subjects) > 0 {
		if manifests.ClusterRoleBinding.Subjects[0].Namespace != namespace {
			t.Errorf("ClusterRoleBinding subject namespace: expected %s, got %s",
				namespace, manifests.ClusterRoleBinding.Subjects[0].Namespace)
		}
	}

	// Verify DaemonSet
	if manifests.DaemonSet == nil {
		t.Error("DaemonSet is nil")
	} else {
		if manifests.DaemonSet.Namespace != namespace {
			t.Errorf("DaemonSet namespace: expected %s, got %s", namespace, manifests.DaemonSet.Namespace)
		}
		// Verify the image is correctly set
		if len(manifests.DaemonSet.Spec.Template.Spec.Containers) > 0 {
			if manifests.DaemonSet.Spec.Template.Spec.Containers[0].Image != image {
				t.Errorf("DaemonSet image: expected %s, got %s", image, manifests.DaemonSet.Spec.Template.Spec.Containers[0].Image)
			}
		} else {
			t.Error("DaemonSet has no containers")
		}
	}

	// Verify ValidatingAdmissionPolicy
	if manifests.ValidatingAdmissionPolicy == nil {
		t.Error("ValidatingAdmissionPolicy is nil")
	} else {
		// Verify the namespace was correctly set in the matchConditions expression
		foundMatchCondition := false
		for _, mc := range manifests.ValidatingAdmissionPolicy.Spec.MatchConditions {
			if mc.Name == "isRestrictedUser" {
				foundMatchCondition = true
				expectedExpr := `request.userInfo.username == "system:serviceaccount:test-namespace:dra-cpu-driver-service-account"`
				if mc.Expression != expectedExpr {
					t.Errorf("ValidatingAdmissionPolicy expression: expected\n%s\ngot\n%s", expectedExpr, mc.Expression)
				}
			}
		}
		if !foundMatchCondition {
			t.Error("ValidatingAdmissionPolicy: did not find isRestrictedUser matchCondition")
		}
	}

	// Verify ValidatingAdmissionPolicyBinding
	if manifests.ValidatingAdmissionPolicyBinding == nil {
		t.Error("ValidatingAdmissionPolicyBinding is nil")
	}
}

func TestGetServiceAccount(t *testing.T) {
	namespace := "test-namespace"
	sa, err := GetServiceAccount(namespace)
	if err != nil {
		t.Fatalf("GetServiceAccount failed: %v", err)
	}

	if sa == nil {
		t.Fatal("GetServiceAccount returned nil")
	}

	if sa.Namespace != namespace {
		t.Errorf("Expected namespace %s, got %s", namespace, sa.Namespace)
	}

	if sa.Name == "" {
		t.Error("ServiceAccount name is empty")
	}
}

func TestGetClusterRole(t *testing.T) {
	cr, err := GetClusterRole()
	if err != nil {
		t.Fatalf("GetClusterRole failed: %v", err)
	}

	if cr == nil {
		t.Fatal("GetClusterRole returned nil")
	}

	if cr.Name == "" {
		t.Error("ClusterRole name is empty")
	}

	if len(cr.Rules) == 0 {
		t.Error("ClusterRole has no rules")
	}
}

func TestGetClusterRoleBinding(t *testing.T) {
	namespace := "test-namespace"
	crb, err := GetClusterRoleBinding(namespace)
	if err != nil {
		t.Fatalf("GetClusterRoleBinding failed: %v", err)
	}

	if crb == nil {
		t.Fatal("GetClusterRoleBinding returned nil")
	}

	if crb.Name == "" {
		t.Error("ClusterRoleBinding name is empty")
	}

	if len(crb.Subjects) == 0 {
		t.Error("ClusterRoleBinding has no subjects")
	}

	// Verify namespace is set correctly
	for _, subject := range crb.Subjects {
		if subject.Namespace != namespace {
			t.Errorf("Expected subject namespace %s, got %s", namespace, subject.Namespace)
		}
	}
}

func TestGetDaemonSet(t *testing.T) {
	namespace := "test-namespace"
	image := "quay.io/test/test-image:v1.0.0"
	ds, err := GetDaemonSet(namespace, image)
	if err != nil {
		t.Fatalf("GetDaemonSet failed: %v", err)
	}

	if ds == nil {
		t.Fatal("GetDaemonSet returned nil")
	}

	if ds.Namespace != namespace {
		t.Errorf("Expected namespace %s, got %s", namespace, ds.Namespace)
	}

	if ds.Name == "" {
		t.Error("DaemonSet name is empty")
	}

	if len(ds.Spec.Template.Spec.Containers) == 0 {
		t.Error("DaemonSet has no containers")
	} else {
		// Verify the image is correctly set
		if ds.Spec.Template.Spec.Containers[0].Image != image {
			t.Errorf("Expected image %s, got %s", image, ds.Spec.Template.Spec.Containers[0].Image)
		}
	}
}

func TestGetDeviceClasses(t *testing.T) {
	deviceClasses, err := GetDeviceClasses()
	if err != nil {
		t.Fatalf("GetDeviceClasses failed: %v", err)
	}

	if len(deviceClasses) != 2 {
		t.Errorf("Expected 2 DeviceClasses, got %d", len(deviceClasses))
	}

	// Check that we have both exclusive-cpu and shared-cpu
	hasExclusive := false
	hasShared := false
	for _, dc := range deviceClasses {
		if dc.Name == "exclusive-cpu" {
			hasExclusive = true
		}
		if dc.Name == "shared-cpu" {
			hasShared = true
		}
	}

	if !hasExclusive {
		t.Error("Missing exclusive-cpu DeviceClass")
	}
	if !hasShared {
		t.Error("Missing shared-cpu DeviceClass")
	}
}

func TestGetExclusiveCPUDeviceClass(t *testing.T) {
	dc, err := GetExclusiveCPUDeviceClass()
	if err != nil {
		t.Fatalf("GetExclusiveCPUDeviceClass failed: %v", err)
	}

	if dc == nil {
		t.Fatal("GetExclusiveCPUDeviceClass returned nil")
	}

	if dc.Name != "exclusive-cpu" {
		t.Errorf("Expected name 'exclusive-cpu', got '%s'", dc.Name)
	}
}

func TestGetSharedCPUDeviceClass(t *testing.T) {
	dc, err := GetSharedCPUDeviceClass()
	if err != nil {
		t.Fatalf("GetSharedCPUDeviceClass failed: %v", err)
	}

	if dc == nil {
		t.Fatal("GetSharedCPUDeviceClass returned nil")
	}

	if dc.Name != "shared-cpu" {
		t.Errorf("Expected name 'shared-cpu', got '%s'", dc.Name)
	}
}

func TestGetValidatingAdmissionPolicy(t *testing.T) {
	namespace := "test-namespace"
	vap, err := GetValidatingAdmissionPolicy(namespace)
	if err != nil {
		t.Fatalf("GetValidatingAdmissionPolicy failed: %v", err)
	}

	if vap == nil {
		t.Fatal("GetValidatingAdmissionPolicy returned nil")
	}

	if vap.Name == "" {
		t.Error("ValidatingAdmissionPolicy name is empty")
	}

	// Verify the namespace was correctly set in the matchConditions expression
	foundMatchCondition := false
	for _, mc := range vap.Spec.MatchConditions {
		if mc.Name == "isRestrictedUser" {
			foundMatchCondition = true
			expectedExpr := `request.userInfo.username == "system:serviceaccount:test-namespace:dra-cpu-driver-service-account"`
			if mc.Expression != expectedExpr {
				t.Errorf("Expected expression:\n%s\nGot:\n%s", expectedExpr, mc.Expression)
			}
		}
	}
	if !foundMatchCondition {
		t.Error("Did not find isRestrictedUser matchCondition")
	}
}

func TestGetValidatingAdmissionPolicyBinding(t *testing.T) {
	vapb, err := GetValidatingAdmissionPolicyBinding()
	if err != nil {
		t.Fatalf("GetValidatingAdmissionPolicyBinding failed: %v", err)
	}

	if vapb == nil {
		t.Fatal("GetValidatingAdmissionPolicyBinding returned nil")
	}

	if vapb.Name == "" {
		t.Error("ValidatingAdmissionPolicyBinding name is empty")
	}

	if vapb.Spec.PolicyName == "" {
		t.Error("ValidatingAdmissionPolicyBinding has no policy name")
	}
}
