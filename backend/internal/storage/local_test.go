package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorage_PutAndGet(t *testing.T) {
	dir := t.TempDir()
	ls, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}
	ctx := context.Background()
	content := "hello world"
	key := "photos/abc/thumb.webp"
	if err := ls.Put(ctx, key, strings.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	rc, err := ls.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != content {
		t.Errorf("got %q, want %q", string(data), content)
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	ls, _ := NewLocalStorage(dir)
	ctx := context.Background()
	key := "photos/abc/thumb.webp"
	ls.Put(ctx, key, strings.NewReader("data"))
	if err := ls.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := ls.Get(ctx, key)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestLocalStorage_URL(t *testing.T) {
	dir := t.TempDir()
	ls, _ := NewLocalStorage(dir)
	url, err := ls.URL(context.Background(), "photos/abc/thumb.webp")
	if err != nil {
		t.Fatalf("URL: %v", err)
	}
	if url != "/media/photos/abc/thumb.webp" {
		t.Errorf("got %q, want %q", url, "/media/photos/abc/thumb.webp")
	}
}

func TestLocalStorage_PutCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	ls, _ := NewLocalStorage(dir)
	key := "deeply/nested/path/file.webp"
	if err := ls.Put(context.Background(), key, strings.NewReader("data")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	fullPath := filepath.Join(dir, key)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Error("expected file to exist at nested path")
	}
}

func TestLocalStorage_GetNotFound(t *testing.T) {
	dir := t.TempDir()
	ls, _ := NewLocalStorage(dir)
	_, err := ls.Get(context.Background(), "nonexistent/file.webp")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
