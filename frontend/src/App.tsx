import React, { useState, useEffect } from 'react';
import { StartVless, StopVless, GetProfiles, DeleteProfile, TcpPing, GetSettings, SaveSettings, GetRunningState, UrlTest, GetLogs, CheckAppUpdate } from "../wailsjs/go/main/App";
import { EventsOn, EventsOff, WindowMinimise, Quit, WindowToggleMaximise } from "../wailsjs/runtime/runtime";
import { main } from "../wailsjs/go/models";

import { Sidebar } from './views/Sidebar';
import { Dashboard } from './views/Dashboard';
import { SettingsView } from './views/Settings';
import { RoutingView } from './views/Routing';
import { LogsView } from './views/Logs';
import { UpdateNotification } from './components/UpdateNotification';

type UIProfile = main.Profile & { latency?: number };
interface TrafficData { up: number; down: number; }

interface UpdateInfo {
    available: boolean;
    version: string;
    current_ver: string;
    release_url: string;
    body: string;
}

function App() {
    const [view, setView] = useState<"dashboard" | "settings" | "logs" | "routing">("dashboard");
    const [status, setStatus] = useState("Ready");
    const [connectionState, setConnectionState] = useState<"disconnected" | "connecting" | "connected">("disconnected");
    const [traffic, setTraffic] = useState<TrafficData>({ up: 0, down: 0 });
    const [profiles, setProfiles] = useState<UIProfile[]>([]);
    const [settings, setSettingsState] = useState<main.Settings>(new main.Settings());
    const [activeSettings, setActiveSettings] = useState<main.Settings>(new main.Settings());
    const [selectedId, setSelectedId] = useState<string | null>(null);
    const [logs, setLogs] = useState<string[]>([]);
    const [isPinging, setIsPinging] = useState(false);
    
    const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null);
    const [showUpdate, setShowUpdate] = useState(false);

    const [errorMsg, setErrorMsg] = useState<string | null>(null);

    const stripAnsi = (str: string) => str.replace(/\x1b\[[0-9;]*m/g, '');
    const hasChanges = JSON.stringify(settings) !== JSON.stringify(activeSettings);

    useEffect(() => {
        initApp();
        EventsOn("log", (msg: string) => setLogs(p => [...p, stripAnsi(msg)].slice(-500)));
        EventsOn("traffic", (jsonStr: string) => { try { setTraffic(JSON.parse(jsonStr)); } catch (e) {} });
        EventsOn("error", (msg: string) => {
            setErrorMsg(msg);
            setTimeout(() => setErrorMsg(null), 5000);
        });
        EventsOn("connection_lost", (msg: string) => {
            setConnectionState("disconnected"); setStatus(msg || "Crashed"); setTraffic({ up: 0, down: 0 });
            setErrorMsg(msg || "Connection lost unexpectedly");
        });
        EventsOn("connection_status", (status: string) => {
            if (status === "connected") {
                setConnectionState("connected");
                setStatus("Secured");
                setActiveSettings(prev => prev);
            } else {
                setConnectionState("disconnected");
                setStatus("Disconnected");
                setTraffic({ up: 0, down: 0 });
            }
        });

        return () => {
            EventsOff("log");
            EventsOff("traffic");
            EventsOff("error");
            EventsOff("connection_lost");
            EventsOff("connection_status");
        };
    }, []);

    const initApp = async () => {
        const s = await GetSettings();
        setSettingsState(s);
        setActiveSettings(s);

        const oldLogs = await GetLogs();
        if (oldLogs && oldLogs.length > 0) {
            setLogs(oldLogs.map(stripAnsi));
        }

        CheckAppUpdate().then((info: any) => {
            if (info && info.available) {
                setUpdateInfo(info);
                setShowUpdate(true);
            }
        });

        await refreshProfiles();
        try {
            const isRunning = await GetRunningState();
            if (isRunning) {
                setConnectionState("connected");
                setStatus("Secured");
            }
        } catch (e) {
            console.error("GetRunningState failed", e);
        }
    };

    const refreshProfiles = async () => {
        const list = await GetProfiles();
        const uiList = (list || []) as UIProfile[];
        setProfiles(uiList);
        if (uiList.length > 0 && !selectedId) setSelectedId(uiList[0].id);
    };

    const checkPings = async (list: UIProfile[]) => {
        if (!list || list.length === 0) return;
        setIsPinging(true);
        const updatedList = [...list];
        for(let i=0; i<updatedList.length; i++) updatedList[i].latency = undefined;
        setProfiles([...updatedList]);

        for (let i = 0; i < updatedList.length; i++) {
            const ms = await UrlTest(updatedList[i].id);
            updatedList[i].latency = ms;
            setProfiles([...updatedList]);
        }
        setIsPinging(false);
    };

    const toggleConnection = async () => {
        if (connectionState === "disconnected") {
            const currentProfile = profiles.find(p => p.id === selectedId);
            if (!currentProfile) { setStatus("Select profile"); return; }

            setConnectionState("connecting");
            setStatus("Starting...");

            const res = await StartVless(currentProfile.key);
            if (res === "Connected") {
                setConnectionState("connected"); setStatus("Secured");
                setActiveSettings(settings);
            } else {
                setConnectionState("disconnected"); setStatus(res);
                setErrorMsg(res);
            }
        } else {
            setStatus("Stopping..."); await StopVless();
            setStatus("Disconnected"); setConnectionState("disconnected"); setTraffic({ up: 0, down: 0 });
        }
    };

    const handleRestart = async () => {
        if (connectionState !== "connected") return;
        setStatus("Restarting..."); setConnectionState("connecting");
        await StopVless();
        const currentProfile = profiles.find(p => p.id === selectedId);
        if (currentProfile) {
            const res = await StartVless(currentProfile.key);
            if (res === "Connected") {
                setConnectionState("connected"); setStatus("Secured");
                setActiveSettings(settings);
            } else {
                setConnectionState("disconnected"); setStatus(res);
            }
        } else { setConnectionState("disconnected"); }
    };

    const handleSelectProfile = async (id: string) => {
        setSelectedId(id);
        if (connectionState === "connected") {
            setStatus("Switching..."); await StopVless();
            const profile = profiles.find(p => p.id === id);
            if (profile) {
                const res = await StartVless(profile.key);
                if (res === "Connected") { setConnectionState("connected"); setStatus("Secured"); setActiveSettings(settings); }
                else { setConnectionState("disconnected"); setStatus(res); setErrorMsg(res); }
            }
        }
    };

    const handleDelete = async (e: React.MouseEvent, id: string) => {
        e.stopPropagation(); await DeleteProfile(id); refreshProfiles();
    };

    const handleSettingsUpdate = async (newSettings: main.Settings) => {
        setSettingsState(newSettings);
        await SaveSettings(new main.Settings(newSettings));
    };

    return (
        <div className="h-screen w-full flex bg-[#09090b] font-sans text-white overflow-hidden relative selection:bg-purple-500/30">
            <Sidebar view={view} setView={setView} />
            <div className="flex-1 flex flex-col relative">
                <div className="h-10 flex justify-end items-center px-4 gap-3 bg-[#09090b] z-50 relative" style={{ "--wails-draggable": "drag" } as React.CSSProperties}>
                    <button onClick={WindowMinimise} className="text-gray-500 hover:text-white p-1" style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}><svg className="w-4 h-4" viewBox="0 0 24 24" stroke="currentColor" fill="none"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4"/></svg></button>
                    <button onClick={WindowToggleMaximise} className="text-gray-500 hover:text-white p-1" style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}><svg className="w-3 h-3" viewBox="0 0 24 24" stroke="currentColor" fill="none"><rect width="18" height="18" x="3" y="3" rx="2" ry="2" strokeWidth="2"/></svg></button>
                    <button onClick={Quit} className="text-gray-500 hover:text-red-400 p-1" style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}><svg className="w-4 h-4" viewBox="0 0 24 24" stroke="currentColor" fill="none"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12"/></svg></button>
                </div>

                <div className="flex-1 p-8 flex items-center justify-center overflow-hidden relative">
                    <div className={`absolute top-4 left-1/2 -translate-x-1/2 z-50 transition-all duration-300 ${errorMsg ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-4 pointer-events-none"}`}>
                        <div className="bg-red-500/10 border border-red-500/50 text-red-400 px-4 py-3 rounded-xl shadow-lg backdrop-blur-md flex items-center gap-3">
                            <svg className="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                            <span className="text-xs font-bold">{errorMsg}</span>
                            <button onClick={() => setErrorMsg(null)} className="ml-2 hover:text-white"><svg className="w-4 h-4" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12"/></svg></button>
                        </div>
                    </div>

                    {view === "dashboard" && (
                        <Dashboard
                            profiles={profiles} selectedId={selectedId} status={status}
                            connectionState={connectionState} traffic={traffic} isPinging={isPinging}
                            onToggle={toggleConnection} onSelect={handleSelectProfile}
                            onDelete={handleDelete} onPing={(e) => { e.stopPropagation(); checkPings(profiles); }}
                            onRefreshProfiles={refreshProfiles}
                        />
                    )}
                    {view === "settings" && (
                        <SettingsView
                            settings={settings} onUpdate={handleSettingsUpdate}
                            hasChanges={hasChanges} isRunning={connectionState === "connected"} onRestart={handleRestart}
                        />
                    )}
                    {view === "routing" && (
                        <RoutingView
                            settings={settings} onUpdate={handleSettingsUpdate}
                            hasChanges={hasChanges} isRunning={connectionState === "connected"} onRestart={handleRestart}
                        />
                    )}
                    {view === "logs" && <LogsView logs={logs} onClear={() => setLogs([])} />}
                </div>

                <UpdateNotification 
                    visible={showUpdate && !!updateInfo?.available} 
                    version={updateInfo?.version || ""} 
                    url={updateInfo?.release_url || ""} 
                    onClose={() => setShowUpdate(false)}
                />
            </div>
        </div>
    )
}

export default App