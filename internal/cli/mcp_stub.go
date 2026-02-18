package cli

import (
	mcpserver "github.com/rzzdr/marrow/internal/mcp"
	"github.com/rzzdr/marrow/internal/store"
)

func runMCPServer(s *store.Store) error {
	return mcpserver.Serve(s)
}
