package k8smodule

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

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
