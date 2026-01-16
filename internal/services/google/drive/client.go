package drive

import (
	"context"
	"fmt"

	"google.golang.org/api/drive/v3"

	googleauth "github.com/zepzeper/vulgar/internal/services/google"
)

type Client struct {
	svc *drive.Service
}

type File struct {
	ID           string
	Name         string
	MimeType     string
	Size         int64
	ModifiedTime string
	WebViewLink  string
	Parents      []string
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := googleauth.NewDriveService(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{svc: svc}, nil
}

func (c *Client) ListFiles(ctx context.Context, folderID string, limit int64) ([]File, error) {
	query := c.svc.Files.List().
		PageSize(limit).
		Fields("files(id, name, mimeType, size, modifiedTime, webViewLink, parents)")

	if folderID != "" {
		query = query.Q(fmt.Sprintf("'%s' in parents", folderID))
	}

	result, err := query.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list files failed: %w", err)
	}

	files := make([]File, len(result.Files))
	for i, f := range result.Files {
		files[i] = File{
			ID:           f.Id,
			Name:         f.Name,
			MimeType:     f.MimeType,
			Size:         f.Size,
			ModifiedTime: f.ModifiedTime,
			WebViewLink:  f.WebViewLink,
			Parents:      f.Parents,
		}
	}

	return files, nil
}

func (c *Client) GetFile(ctx context.Context, fileID string) (*File, error) {
	f, err := c.svc.Files.Get(fileID).
		Fields("id, name, mimeType, size, modifiedTime, webViewLink, parents").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get file failed: %w", err)
	}

	return &File{
		ID:           f.Id,
		Name:         f.Name,
		MimeType:     f.MimeType,
		Size:         f.Size,
		ModifiedTime: f.ModifiedTime,
		WebViewLink:  f.WebViewLink,
		Parents:      f.Parents,
	}, nil
}

func (c *Client) SearchFiles(ctx context.Context, query string, limit int64) ([]File, error) {
	q := fmt.Sprintf("name contains '%s'", query)

	result, err := c.svc.Files.List().
		Q(q).
		PageSize(limit).
		Fields("files(id, name, mimeType, size, modifiedTime, webViewLink, parents)").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("search files failed: %w", err)
	}

	files := make([]File, len(result.Files))
	for i, f := range result.Files {
		files[i] = File{
			ID:           f.Id,
			Name:         f.Name,
			MimeType:     f.MimeType,
			Size:         f.Size,
			ModifiedTime: f.ModifiedTime,
			WebViewLink:  f.WebViewLink,
			Parents:      f.Parents,
		}
	}

	return files, nil
}
