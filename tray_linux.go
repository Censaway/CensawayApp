//go:build linux

package main

/*
#cgo pkg-config: ayatana-appindicator3-0.1 gtk+-3.0
#include <libayatana-appindicator/app-indicator.h>
#include <gtk/gtk.h>
#include <stdlib.h>

static AppIndicator *indicator;
static GtkWidget *menu;
static GtkWidget *item_show;
static GtkWidget *item_quit;

extern void goOnShow();
extern void goOnQuit();

static void on_show_activate(GtkMenuItem *m, gpointer data) { goOnShow(); }
static void on_quit_activate(GtkMenuItem *m, gpointer data) { goOnQuit(); }

static gboolean create_tray_callback(gpointer data) {
    char *icon_path = (char*)data;

    indicator = app_indicator_new("censaway-tray", icon_path, APP_INDICATOR_CATEGORY_APPLICATION_STATUS);

    app_indicator_set_status(indicator, APP_INDICATOR_STATUS_ACTIVE);
    app_indicator_set_title(indicator, "Censaway");

    menu = gtk_menu_new();

    item_show = gtk_menu_item_new_with_label("Show Window");
    g_signal_connect(item_show, "activate", G_CALLBACK(on_show_activate), NULL);
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), item_show);

    GtkWidget *sep = gtk_separator_menu_item_new();
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), sep);

    item_quit = gtk_menu_item_new_with_label("Quit");
    g_signal_connect(item_quit, "activate", G_CALLBACK(on_quit_activate), NULL);
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), item_quit);

    gtk_widget_show_all(menu);
    app_indicator_set_menu(indicator, GTK_MENU(menu));

    g_free(icon_path);
    return FALSE;
}

static void schedule_tray_creation(char* icon_path) {
    g_idle_add(create_tray_callback, g_strdup(icon_path));
}
*/
import "C"

import (
	"context"
	"os"
	"path/filepath"
	"unsafe"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var trayCtx context.Context
var trayApp *App

func (a *App) SetupTray(ctx context.Context) {
	trayCtx = ctx
	trayApp = a

	iconPath := filepath.Join(os.TempDir(), "censaway-tray.png")

	if len(a.Icon) > 0 {
		os.WriteFile(iconPath, a.Icon, 0644)
	}

	cIconPath := C.CString(iconPath)
	defer C.free(unsafe.Pointer(cIconPath))

	C.schedule_tray_creation(cIconPath)
}

func (a *App) OnExit() {
}

func (a *App) updateTrayState(connected bool) {
}

//export goOnShow
func goOnShow() {
	wailsRuntime.WindowShow(trayCtx)
	wailsRuntime.WindowUnminimise(trayCtx)
}

//export goOnQuit
func goOnQuit() {
	trayApp.isQuitting = true
	trayApp.StopVless()
	wailsRuntime.Quit(trayCtx)
}