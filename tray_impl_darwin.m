#include "tray_impl_darwin.h"
#include <Cocoa/Cocoa.h>
#include "_cgo_export.h"

const int ID_STATUS = 0;
const int ID_CONNECT = 1;
const int ID_SHOW = 2;
const int ID_QUIT = 3;

@interface CensawayTrayHandlerV3 : NSObject
@end

static NSStatusItem *statusItem;
static CensawayTrayHandlerV3 *handler;
static NSMenuItem *statusMenuItem;
static NSMenuItem *connectMenuItem;

@implementation CensawayTrayHandlerV3
- (void)menuItemClicked:(id)sender {
    NSMenuItem *item = (NSMenuItem *)sender;
    goOnTrayClick((int)[item tag]);
}
@end

void update_tray_state_c(int connected) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (connected) {
            [connectMenuItem setTitle:@"Disconnect"];
            [statusMenuItem setTitle:@"Status: Connected"];
            if (statusItem.button) {
                [statusItem.button setToolTip:@"Censaway: Connected"];
            }
        } else {
            [connectMenuItem setTitle:@"Connect"];
            [statusMenuItem setTitle:@"Status: Disconnected"];
            if (statusItem.button) {
                [statusItem.button setToolTip:@"Censaway: Disconnected"];
            }
        }
    });
}

NSImage* resizeImage(NSImage* sourceImage, NSSize newSize) {
    if (!sourceImage.isValid) return nil;

    NSImage *smallImage = [[NSImage alloc] initWithSize: newSize];
    [smallImage lockFocus];
    [sourceImage setSize: newSize];
    [[NSGraphicsContext currentContext] setImageInterpolation:NSImageInterpolationHigh];
    [sourceImage drawAtPoint:NSZeroPoint fromRect:CGRectMake(0, 0, newSize.width, newSize.height) operation:NSCompositingOperationCopy fraction:1.0];
    [smallImage unlockFocus];
    return smallImage;
}

void init_tray_c(void* data, int length) {
    dispatch_async(dispatch_get_main_queue(), ^{
        handler = [[CensawayTrayHandlerV3 alloc] init];
        statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        [statusItem retain];

        NSData *imageData = [NSData dataWithBytes:data length:length];
        NSImage *originalImage = [[NSImage alloc] initWithData:imageData];
        
        if (originalImage) {
            NSImage *resized = resizeImage(originalImage, NSMakeSize(22, 22));
            
            [resized setTemplate:NO]; 
            
            [statusItem.button setImage:resized];
        } else {
             NSImage *fallback = [NSImage imageWithSystemSymbolName:@"network" accessibilityDescription:nil];
             [statusItem.button setImage:fallback];
        }
        
        [statusItem.button setToolTip:@"Censaway"];

        NSMenu *menu = [[NSMenu alloc] init];

        statusMenuItem = [menu addItemWithTitle:@"Status: Disconnected" action:nil keyEquivalent:@""];
        [statusMenuItem setTag:ID_STATUS];
        [statusMenuItem setEnabled:NO];

        [menu addItem:[NSMenuItem separatorItem]];

        connectMenuItem = [menu addItemWithTitle:@"Connect" action:@selector(menuItemClicked:) keyEquivalent:@""];
        [connectMenuItem setTarget:handler];
        [connectMenuItem setTag:ID_CONNECT];

        NSMenuItem *showItem = [menu addItemWithTitle:@"Show Window" action:@selector(menuItemClicked:) keyEquivalent:@""];
        [showItem setTarget:handler];
        [showItem setTag:ID_SHOW];

        [menu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *quitItem = [menu addItemWithTitle:@"Quit" action:@selector(menuItemClicked:) keyEquivalent:@""];
        [quitItem setTarget:handler];
        [quitItem setTag:ID_QUIT];

        [statusItem setMenu:menu];
    });
}