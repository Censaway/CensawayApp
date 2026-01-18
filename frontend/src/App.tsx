import React, { useEffect } from 'react';
import { GetProfiles, GetSettings, GetLogs, CheckAppUpdate, GetRunningState, GetAppVersion } from "../wailsjs/go/main/App";
import { EventsOn, EventsOff, WindowMinimise, Quit, WindowToggleMaximise } from "../wailsjs/runtime/runtime";
import { useAppStore, UIProfile, AppState } from './store/appStore';
import { main } from "../wailsjs/go/models";

import { Sidebar } from './views/Sidebar';
import { Dashboard } from './views/Dashboard';
import { SettingsView } from './views/Settings';
import { RoutingView } from './views/Routing';
import { LogsView } from './views/Logs';
import { UpdateNotification } from './components/UpdateNotification';

function App() {
    const { 
        view, errorMsg, 
        addLog, setTraffic, setConnectionState, setStatus, setErrorMsg,
        setSettings, setLogs, setProfiles, setSelectedId, setUpdateInfo, setShowUpdate,
        setCurrentIp, setAppVersion
    } = useAppStore();

    useEffect(() => {
        initApp();

        EventsOn("log", (msg: string) => addLog(msg));
        EventsOn("traffic", (jsonStr: string) => { try { setTraffic(JSON.parse(jsonStr)); } catch (e) {} });
        EventsOn("error", (msg: string) => {
            setErrorMsg(msg);
            setTimeout(() => setErrorMsg(null), 5000);
        });
        EventsOn("connection_lost", (msg: string) => {
            setConnectionState("disconnected"); 
            setStatus(msg || "Crashed"); 
            setTraffic({ up: 0, down: 0 });
            setCurrentIp(null);
            setErrorMsg(msg || "Connection lost unexpectedly");
        });
        EventsOn("connection_status", (status: string) => {
            if (status === "connected") {
                setConnectionState("connected");
                setStatus("Secured");
            } else {
                setConnectionState("disconnected");
                setStatus("Disconnected");
                setTraffic({ up: 0, down: 0 });
                setCurrentIp(null);
            }
        });
        EventsOn("update_available", (info: any) => {
             setUpdateInfo(info);
             setShowUpdate(true);
        });

        return () => {
            EventsOff("log");
            EventsOff("traffic");
            EventsOff("error");
            EventsOff("connection_lost");
            EventsOff("connection_status");
            EventsOff("update_available");
        };
    }, []);

    const initApp = async () => {
        try {
            const ver = await GetAppVersion();
            setAppVersion(ver);

            const s = await GetSettings();
            setSettings(s);

            const oldLogs = await GetLogs();
            if (oldLogs && oldLogs.length > 0) {
                const stripAnsi = (str: string) => str.replace(/\x1b\[[0-9;]*m/g, '');
                setLogs(oldLogs.map(stripAnsi));
            }

            const info: any = await CheckAppUpdate();
            if (info && info.available) {
                setUpdateInfo(info);
                setShowUpdate(true);
            }

            const list = await GetProfiles();
            const uiList = (list || []) as UIProfile[];
            setProfiles(uiList);
            if (uiList.length > 0 && s.last_profile_id) {
                 const exists = uiList.find((p: UIProfile) => p.id === s.last_profile_id);
                 setSelectedId(exists ? exists.id : uiList[0].id);
            } else if (uiList.length > 0) {
                 setSelectedId(uiList[0].id);
            }

            const isRunning = await GetRunningState();
            if (isRunning) {
                setConnectionState("connected");
                setStatus("Secured");
            }
        } catch(e) {
            console.error("Init failed", e);
        }
    };

    const updateInfo = useAppStore((state: AppState) => state.updateInfo);
    const showUpdate = useAppStore((state: AppState) => state.showUpdate);

    return (
        <div className="h-screen w-full flex bg-[#09090b] font-sans text-white overflow-hidden relative selection:bg-purple-500/30">
            <Sidebar />
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

                    {view === "dashboard" && <Dashboard />}
                    {view === "settings" && <SettingsView />}
                    {view === "routing" && <RoutingView />}
                    {view === "logs" && <LogsView />}
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