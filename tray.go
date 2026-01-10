package main

import (
	"context"
)

type TrayHandler interface {
	SetupTray(ctx context.Context)
	OnExit()
}
