package gdrive

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	gauth "github.com/zepzeper/vulgar/internal/modules/integrations/google"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/drive/v3"
)

// luaConfigure creates a new Google Drive client using OAuth authentication
// Usage: local client, err = gdrive.configure()
// Note: Requires prior authentication via 'vulgar gdrive login'
func luaConfigure(L *lua.LState) int {
	ctx := context.Background()

	clientOpt, err := gauth.ClientOption(ctx)
	if err != nil {
		return util.PushError(L, "failed to get OAuth credentials: %v (run 'vulgar gdrive login' first)", err)
	}

	// Create Drive service
	service, err := drive.NewService(ctx, clientOpt)
	if err != nil {
		return util.PushError(L, "failed to create drive service: %v", err)
	}

	// Wrap and return client
	wrapper := &clientWrapper{
		service: service,
		ctx:     ctx,
	}

	L.Push(registerClient(L, wrapper))
	L.Push(lua.LNil)
	return 2
}
