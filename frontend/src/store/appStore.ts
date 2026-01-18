import { create } from 'zustand';
import { main } from '../../wailsjs/go/models';

export interface TrafficData { up: number; down: number; }
export type UIProfile = main.Profile & { latency?: number };

interface UpdateInfo {
    available: boolean;
    version: string;
    current_ver: string;
    release_url: string;
    body: string;
}

export interface AppState {
    view: "dashboard" | "settings" | "logs" | "routing";
    connectionState: "disconnected" | "connecting" | "connected";
    status: string;
    traffic: TrafficData;
    currentIp: string | null;
    profiles: UIProfile[];
    selectedId: string | null;
    settings: main.Settings;
    logs: string[];
    isPinging: boolean;
    updateInfo: UpdateInfo | null;
    showUpdate: boolean;
    errorMsg: string | null;
    appVersion: string;

    setView: (v: "dashboard" | "settings" | "logs" | "routing") => void;
    setConnectionState: (s: "disconnected" | "connecting" | "connected") => void;
    setStatus: (s: string) => void;
    setTraffic: (t: TrafficData) => void;
    setCurrentIp: (ip: string | null) => void;
    setProfiles: (p: UIProfile[]) => void;
    setSelectedId: (id: string | null) => void;
    setSettings: (s: main.Settings) => void;
    addLog: (msg: string) => void;
    setLogs: (logs: string[]) => void;
    clearLogs: () => void;
    setIsPinging: (b: boolean) => void;
    setUpdateInfo: (u: UpdateInfo | null) => void;
    setShowUpdate: (b: boolean) => void;
    setErrorMsg: (msg: string | null) => void;
    setAppVersion: (v: string) => void;
}

export const useAppStore = create<AppState>((set) => ({
    view: "dashboard",
    connectionState: "disconnected",
    status: "Ready",
    traffic: { up: 0, down: 0 },
    currentIp: null,
    profiles: [],
    selectedId: null,
    settings: new main.Settings(),
    logs: [],
    isPinging: false,
    updateInfo: null,
    showUpdate: false,
    errorMsg: null,
    appVersion: "v0.0.0",

    setView: (view) => set({ view }),
    setConnectionState: (connectionState) => set({ connectionState }),
    setStatus: (status) => set({ status }),
    setTraffic: (traffic) => set({ traffic }),
    setCurrentIp: (currentIp) => set({ currentIp }),
    setProfiles: (profiles) => set({ profiles }),
    setSelectedId: (selectedId) => set({ selectedId }),
    setSettings: (settings) => set({ settings }),
    
    addLog: (msg) => set((state) => {
        const stripAnsi = (str: string) => str.replace(/\x1b\[[0-9;]*m/g, '');
        const cleanMsg = stripAnsi(msg);
        const newLogs = [...state.logs, cleanMsg];
        if (newLogs.length > 500) return { logs: newLogs.slice(-500) };
        return { logs: newLogs };
    }),
    setLogs: (logs) => set({ logs }),
    clearLogs: () => set({ logs: [] }),

    setIsPinging: (isPinging) => set({ isPinging }),
    setUpdateInfo: (updateInfo) => set({ updateInfo }),
    setShowUpdate: (showUpdate) => set({ showUpdate }),
    setErrorMsg: (errorMsg) => set({ errorMsg }),
    setAppVersion: (appVersion) => set({ appVersion }),
}));