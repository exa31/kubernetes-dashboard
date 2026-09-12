package k8smodule

import (
	"context"
	"io"
	"log/slog"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func newFakeK8sService(objects ...runtime.Object) (*K8sService, *fake.Clientset) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fakeCS := fake.NewSimpleClientset(objects...)
	cm := &ClientManager{
		Clientset: fakeCS,
		Connected: true,
		Logger:    logger,
		Endpoint:  "https://fake-k8s.cluster.local:6443",
	}
	return NewK8sService(cm), fakeCS
}

func TestK8sService_OfflineReturnsErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	clientMgr := &ClientManager{
		Connected: false,
		Logger:    logger,
	}
	svc := NewK8sService(clientMgr)
	ctx := context.Background()

	// 1. Cluster info should return not connected error
	info, err := svc.GetClusterInfo(ctx)
	if err == nil {
		t.Fatalf("expected error when disconnected, got nil with info: %+v", info)
	}

	// 2. Namespaces should return error
	_, err = svc.ListNamespaces(ctx)
	if err == nil {
		t.Fatal("expected error listing namespaces when disconnected")
	}

	// 3. Secrets should return error
	_, err = svc.ListSecrets(ctx, "default")
	if err == nil {
		t.Fatal("expected error listing secrets when disconnected")
	}

	// 4. Get Secret should return error
	_, err = svc.GetSecret(ctx, "default", "test-secret")
	if err == nil {
		t.Fatal("expected error getting secret when disconnected")
	}

	// 5. Deployments should return error
	_, err = svc.ListDeployments(ctx, "default")
	if err == nil {
		t.Fatal("expected error listing deployments when disconnected")
	}

	// 6. Pods should return error
	_, err = svc.ListPods(ctx, "default")
	if err == nil {
		t.Fatal("expected error listing pods when disconnected")
	}
}

func TestK8sService_FakeClient_Namespaces(t *testing.T) {
	ctx := context.Background()
	ns1 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "production",
			CreationTimestamp: metav1.Now(),
		},
		Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
	}
	svc, _ := newFakeK8sService(ns1)

	// List namespaces
	list, err := svc.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("ListNamespaces failed: %v", err)
	}
	if len(list) != 1 || list[0].Name != "production" {
		t.Fatalf("expected [production], got %+v", list)
	}

	// Create namespace
	created, err := svc.CreateNamespace(ctx, &CreateNamespaceRequest{Name: "staging"})
	if err != nil {
		t.Fatalf("CreateNamespace failed: %v", err)
	}
	if created.Name != "staging" {
		t.Fatalf("expected staging, got %s", created.Name)
	}

	// Verify count is now 2
	list, err = svc.ListNamespaces(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("expected 2 namespaces, got %d (err: %v)", len(list), err)
	}

	// Delete namespace
	err = svc.DeleteNamespace(ctx, "staging")
	if err != nil {
		t.Fatalf("DeleteNamespace failed: %v", err)
	}
}

func TestK8sService_FakeClient_Secrets(t *testing.T) {
	ctx := context.Background()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "api-keys",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"API_KEY": []byte("super-secret-value"),
		},
	}
	svc, _ := newFakeK8sService(secret)

	// List
	list, err := svc.ListSecrets(ctx, "default")
	if err != nil {
		t.Fatalf("ListSecrets failed: %v", err)
	}
	if len(list) != 1 || list[0].Name != "api-keys" {
		t.Fatalf("expected api-keys, got %+v", list)
	}

	// Get detail
	det, err := svc.GetSecret(ctx, "default", "api-keys")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}
	if det.Name != "api-keys" || det.Data["API_KEY"] != "super-secret-value" {
		t.Fatalf("unexpected secret detail: %+v", det)
	}

	// Create new secret
	newSec, err := svc.SaveSecret(ctx, &SaveSecretRequest{
		Name:      "jwt-secret",
		Namespace: "default",
		Type:      string(corev1.SecretTypeOpaque),
		Data:      map[string]string{"TOKEN": "abc123xyz"},
	})
	if err != nil {
		t.Fatalf("SaveSecret failed: %v", err)
	}
	if newSec.Name != "jwt-secret" {
		t.Fatalf("expected jwt-secret, got %s", newSec.Name)
	}

	// Delete secret
	err = svc.DeleteSecret(ctx, "default", "api-keys")
	if err != nil {
		t.Fatalf("DeleteSecret failed: %v", err)
	}
}

func TestK8sService_FakeClient_ConfigMaps(t *testing.T) {
	ctx := context.Background()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "app-env",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Data: map[string]string{
			"APP_ENV": "production",
		},
	}
	svc, _ := newFakeK8sService(cm)

	// List
	list, err := svc.ListConfigMaps(ctx, "default")
	if err != nil {
		t.Fatalf("ListConfigMaps failed: %v", err)
	}
	if len(list) != 1 || list[0].Name != "app-env" {
		t.Fatalf("expected app-env, got %+v", list)
	}

	// Get detail
	det, err := svc.GetConfigMap(ctx, "default", "app-env")
	if err != nil {
		t.Fatalf("GetConfigMap failed: %v", err)
	}
	if det.Data["APP_ENV"] != "production" {
		t.Fatalf("unexpected CM data: %+v", det.Data)
	}

	// Save ConfigMap
	saved, err := svc.SaveConfigMap(ctx, &SaveConfigMapRequest{
		Name:      "feature-flags",
		Namespace: "default",
		Data:      map[string]string{"ENABLE_BETA": "true"},
	})
	if err != nil {
		t.Fatalf("SaveConfigMap failed: %v", err)
	}
	if saved.Name != "feature-flags" {
		t.Fatalf("expected feature-flags, got %s", saved.Name)
	}

	// Delete
	err = svc.DeleteConfigMap(ctx, "default", "app-env")
	if err != nil {
		t.Fatalf("DeleteConfigMap failed: %v", err)
	}
}

func TestK8sService_FakeClient_Deployments(t *testing.T) {
	ctx := context.Background()
	replicas := int32(2)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "backend-api",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
			Labels:            map[string]string{"app": "backend"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "backend"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "backend"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "api", Image: "backend:v1"},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      2,
			ReadyReplicas: 2,
		},
	}
	svc, _ := newFakeK8sService(dep)

	// List deployments
	list, err := svc.ListDeployments(ctx, "default")
	if err != nil {
		t.Fatalf("ListDeployments failed: %v", err)
	}
	if len(list) != 1 || list[0].Name != "backend-api" {
		t.Fatalf("expected backend-api, got %+v", list)
	}

	// Get deployment
	detail, err := svc.GetDeployment(ctx, "default", "backend-api")
	if err != nil {
		t.Fatalf("GetDeployment failed: %v", err)
	}
	if detail.Name != "backend-api" || detail.Replicas != 2 {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	// Restart deployment
	restarted, err := svc.RolloutRestartDeployment(ctx, "default", "backend-api")
	if err != nil {
		t.Fatalf("RolloutRestartDeployment failed: %v", err)
	}
	if restarted.Deployment != "backend-api" {
		t.Fatalf("expected deployment backend-api, got %s", restarted.Deployment)
	}
}

func TestK8sService_FakeClient_Pods(t *testing.T) {
	ctx := context.Background()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "backend-api-xyz-123",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
			Labels:            map[string]string{"app": "backend"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "api", Image: "backend:v1"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "api",
					Ready: true,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{StartedAt: metav1.Now()},
					},
				},
			},
		},
	}
	svc, _ := newFakeK8sService(pod)

	// List pods
	pods, err := svc.ListPods(ctx, "default")
	if err != nil {
		t.Fatalf("ListPods failed: %v", err)
	}
	if len(pods) != 1 || pods[0].Name != "backend-api-xyz-123" {
		t.Fatalf("expected backend pod, got %+v", pods)
	}
	if pods[0].Phase != "Running" {
		t.Fatalf("expected Running phase, got %s", pods[0].Phase)
	}
}

func TestK8sService_FakeClient_Services(t *testing.T) {
	ctx := context.Background()
	svcObj := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "api-service",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: "10.96.0.100",
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 8080},
			},
		},
	}
	svc, _ := newFakeK8sService(svcObj)

	list, err := svc.ListServices(ctx, "default")
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}
	if len(list) != 1 || list[0].Name != "api-service" {
		t.Fatalf("expected api-service, got %+v", list)
	}

	detail, err := svc.GetService(ctx, "default", "api-service")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}
	if detail.ClusterIP != "10.96.0.100" {
		t.Fatalf("expected 10.96.0.100, got %s", detail.ClusterIP)
	}
}

func TestK8sService_FakeClient_CronJobs(t *testing.T) {
	ctx := context.Background()
	isSuspended := false
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "db-cleanup",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "0 2 * * *",
			Suspend:  &isSuspended,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyOnFailure,
							Containers: []corev1.Container{
								{Name: "cleaner", Image: "cleanup:v1"},
							},
						},
					},
				},
			},
		},
	}
	svc, _ := newFakeK8sService(cj)

	list, err := svc.ListCronJobs(ctx, "default")
	if err != nil {
		t.Fatalf("ListCronJobs failed: %v", err)
	}
	if len(list) != 1 || list[0].Name != "db-cleanup" {
		t.Fatalf("expected db-cleanup, got %+v", list)
	}

	// Toggle suspend to true
	suspended, err := svc.ToggleSuspendCronJob(ctx, "default", "db-cleanup")
	if err != nil {
		t.Fatalf("ToggleSuspendCronJob failed: %v", err)
	}
	if !suspended {
		t.Fatalf("expected suspended to be true, got %v", suspended)
	}
}
