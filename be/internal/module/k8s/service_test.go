package k8smodule

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func TestK8sService_OfflineDemoMode(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	clientMgr := &ClientManager{
		Connected: false,
		Logger:    logger,
	}
	svc := NewK8sService(clientMgr)
	ctx := context.Background()

	// 1. Cluster info
	info, err := svc.GetClusterInfo(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.Connected {
		t.Errorf("expected disconnected in demo mode")
	}

	// 2. Namespaces
	nsList, err := svc.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(nsList) == 0 {
		t.Fatalf("expected at least 1 demo namespace")
	}

	// 3. Secrets
	secrets, err := svc.ListSecrets(ctx, "dev-coffe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(secrets) == 0 {
		t.Fatalf("expected secrets in dev-coffe")
	}

	// 4. Get Secret & auto-decoded plaintext data
	detail, err := svc.GetSecret(ctx, "dev-coffe", "be-chat-app-env")
	if err != nil {
		t.Fatalf("expected no error getting secret, got %v", err)
	}
	if detail.Data["APP_NAME"] != "Chat-App" {
		t.Errorf("expected decoded APP_NAME to be 'Chat-App', got '%s'", detail.Data["APP_NAME"])
	}
	if detail.Data["DATABASE_HOST"] != "103.150.226.122" {
		t.Errorf("expected decoded DATABASE_HOST to be '103.150.226.122', got '%s'", detail.Data["DATABASE_HOST"])
	}

	// 5. Save Secret
	detail.Data["APP_VERSION"] = "2.0.0"
	saved, err := svc.SaveSecret(ctx, &SaveSecretRequest{
		Name:      detail.Name,
		Namespace: detail.Namespace,
		Type:      detail.Type,
		Data:      detail.Data,
	})
	if err != nil {
		t.Fatalf("expected no error saving secret, got %v", err)
	}
	if saved.Data["APP_VERSION"] != "2.0.0" {
		t.Errorf("expected APP_VERSION to be updated to '2.0.0'")
	}
}
