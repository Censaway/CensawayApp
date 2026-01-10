import React, { useState, useRef, useEffect } from 'react';
import { AddProfile, CreateSubscription, GetSubscriptions, UpdateSubscription, DeleteSubscription, UpdateProfile } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";
import { ConfirmationModal } from '../components/ConfirmationModal';
import { EditProfileModal } from '../components/EditProfileModal';

type UIProfile = main.Profile & { latency?: number };
interface TrafficData { up: number; down: number; }

interface Props {
    profiles: UIProfile[];
    selectedId: string | null;
    status: string;
    connectionState: "disconnected" | "connecting" | "connected";
    traffic: TrafficData;
    isPinging: boolean;
    onToggle: () => void;
    onSelect: (id: string) => void;
    onDelete: (e: React.MouseEvent, id: string) => void;
    onPing: (e: React.MouseEvent) => void;
    onRefreshProfiles: () => void;
}

export const Dashboard: React.FC<Props> = ({
                                               profiles, selectedId, status, connectionState, traffic, isPinging,
                                               onToggle, onSelect, onDelete, onPing, onRefreshProfiles
                                           }) => {
    const [isAdding, setIsAdding] = useState(false);
    const [addType, setAddType] = useState<"key" | "sub">("key");
    const [inputVal, setInputVal] = useState("");
    const [isProcessing, setIsProcessing] = useState(false);
    const [subscriptions, setSubscriptions] = useState<main.Subscription[]>([]);
    const [updatingSubId, setUpdatingSubId] = useState<string | null>(null);
    const [subToDelete, setSubToDelete] = useState<string | null>(null);
    const [profileToEdit, setProfileToEdit] = useState<UIProfile | null>(null);

    const inputRef = useRef<HTMLInputElement>(null);

    const isRunning = connectionState === "connected";
    const isLoading = connectionState === "connecting";

    useEffect(() => {
        if (isAdding && inputRef.current) {
            setTimeout(() => inputRef.current?.focus(), 100);
        }
    }, [isAdding]);

    React.useEffect(() => { loadSubs(); }, []);
    const loadSubs = async () => { const subs = await GetSubscriptions(); setSubscriptions(subs || []); };
    const formatSpeed = (bytes: number) => { if (bytes === 0) return "0.0 KB/s"; const k = 1024; const sizes = ['B/s', 'KB/s', 'MB/s', 'GB/s']; const i = Math.floor(Math.log(bytes) / Math.log(k)); return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]; };
    const getPingColor = (ms?: number) => { if (ms === undefined) return "bg-gray-600"; if (ms === -1 || ms >= 999) return "bg-red-500"; if (ms < 100) return "bg-green-500"; if (ms < 300) return "bg-yellow-500"; return "bg-orange-500"; };

    const handleAdd = async () => {
        if (!inputVal) return;
        setIsProcessing(true);
        if (addType === "key") { await AddProfile(inputVal); } else { await CreateSubscription(inputVal); await loadSubs(); }
        setInputVal(""); setIsProcessing(false); setIsAdding(false); onRefreshProfiles();
    };

    const handleUpdateSub = async (e: React.MouseEvent, id: string) => {
        e.stopPropagation(); setUpdatingSubId(id); await UpdateSubscription(id); await onRefreshProfiles(); setUpdatingSubId(null);
    };
    const requestDeleteSub = (e: React.MouseEvent, id: string) => { e.stopPropagation(); setSubToDelete(id); };
    const confirmDeleteSub = async () => {
        if (subToDelete) { await DeleteSubscription(subToDelete); await loadSubs(); await onRefreshProfiles(); setSubToDelete(null); }
    };

    const handleSaveProfile = async (name: string, key: string) => {
        if (profileToEdit) {
            await UpdateProfile(profileToEdit.id, name, key);
            onRefreshProfiles();
            setProfileToEdit(null);
        }
    };

    const manualProfiles = profiles.filter(p => !p.subscription_id);
    const getSubProfiles = (subId: string) => profiles.filter(p => p.subscription_id === subId);

    const getServerAddress = (key: string) => {
        try {
            const url = new URL(key);
            return url.hostname;
        } catch (e) {
            return "unknown address";
        }
    };

    const renderProfileItem = (profile: UIProfile) => (
        <div key={profile.id} onClick={() => onSelect(profile.id)} className={`group relative p-3 rounded-xl border transition-all cursor-pointer flex items-center justify-between outline-none ring-0 select-none mb-2 ${selectedId === profile.id ? "bg-purple-500/10 border-purple-500/40" : "bg-transparent border-transparent hover:bg-white/5"} ${isRunning && selectedId !== profile.id ? "opacity-50" : ""}`}>

            <div className="flex items-center gap-3 relative z-10 w-full min-w-0 pr-2">
                <div className={`w-2 h-2 shrink-0 rounded-full transition-all duration-300 ${getPingColor(profile.latency)} ${selectedId === profile.id && !profile.latency ? "shadow-[0_0_8px_#c084fc]" : ""}`}></div>

                <div className="flex flex-col flex-1 overflow-hidden">
                    <div className="flex items-center gap-2 w-full">
                        <span className={`text-xs font-medium truncate transition-colors ${selectedId === profile.id ? "text-white" : "text-gray-400"}`}>
                            {profile.name}
                        </span>

                        {profile.latency !== undefined && (
                            <span className={`text-[9px] font-mono px-1.5 py-0.5 rounded ${profile.latency < 200 ? "bg-green-500/10 text-green-400" : "bg-red-500/10 text-red-400"}`}>
                                {profile.latency >= 999 ? "TIMEOUT" : `${profile.latency}ms`}
                            </span>
                        )}
                    </div>

                    <span className="text-[9px] text-gray-600 font-mono truncate opacity-70">
                        {getServerAddress(profile.key)}
                    </span>
                </div>
            </div>

            <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity duration-200 z-20">
                <div className="absolute inset-0 bg-[#0a0a0e]/80 backdrop-blur-md rounded-lg -z-10 scale-110"></div>
                <button
                    onClick={(e) => { e.stopPropagation(); setProfileToEdit(profile); }}
                    className="text-gray-400 hover:text-white p-1.5 rounded-md hover:bg-white/10 transition-colors relative z-20"
                    title="Edit"
                >
                    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
                </button>
                <button
                    onClick={(e) => onDelete(e, profile.id)}
                    className="text-gray-400 hover:text-red-400 p-1.5 rounded-md hover:bg-red-500/10 transition-colors relative z-20"
                    title="Delete"
                >
                    <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
            </div>
        </div>
    );

    return (
        <>
            <div className="flex w-full max-w-5xl h-[520px] gap-6 animate-[fadeIn_0.3s_ease-out]">
                <div className={`glass flex-1 rounded-3xl overflow-hidden p-8 flex flex-col justify-between relative transition-all duration-700 ${isRunning ? "border-green-500/30 shadow-[0_0_60px_-10px_rgba(34,197,94,0.15)]" : ""}`}>
                    <div className={`absolute top-0 left-0 w-full h-[2px] bg-gradient-to-r transition-all duration-700 ${isRunning ? "from-green-400 via-emerald-400 to-transparent" : "from-purple-500 via-pink-500 to-transparent"}`}></div>
                    <div className="flex justify-between items-center text-gray-500 text-[10px] font-bold tracking-widest uppercase"><span>Status</span><div className="flex items-center gap-2 bg-black/20 px-2 py-1 rounded-full border border-white/5"><span className={isRunning ? "text-green-400" : "text-gray-400"}>{isRunning ? "SECURE" : "IDLE"}</span><div className={`w-1.5 h-1.5 rounded-full ${isRunning ? "bg-green-400 shadow-[0_0_8px_#4ade80]" : "bg-red-500/50"}`}></div></div></div>
                    <div className="flex-1 flex flex-col items-center justify-center gap-6">
                        <button onClick={onToggle} disabled={isLoading || (!profiles || profiles.length === 0)} className={`relative group w-40 h-40 rounded-full flex items-center justify-center outline-none transition-all duration-500 ${isRunning ? "bg-green-500/10 shadow-[0_0_40px_rgba(34,197,94,0.2)]" : "bg-white/5 hover:bg-white/10 shadow-[0_0_40px_rgba(168,85,247,0.1)]"}`}>{isLoading && <div className="absolute inset-[-4px] rounded-full border-2 border-transparent border-t-purple-500 border-r-purple-500/50 animate-spin z-0 pointer-events-none"></div>}<div className={`absolute inset-0 rounded-full border-2 transition-all duration-500 z-10 ${isRunning ? "border-green-500/50 scale-100" : "border-purple-500/30 group-hover:border-purple-400/50 scale-95"}`}></div><div className="z-20"><svg xmlns="http://www.w3.org/2000/svg" className={`h-16 w-16 transition-all duration-500 ${isRunning ? "text-green-400 drop-shadow-[0_0_10px_rgba(74,222,128,0.8)]" : "text-gray-400 group-hover:text-white"}`} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 10V3L4 14h7v7l9-11h-7z" /></svg></div></button>
                        <div className="text-center w-full"><h2 className={`text-3xl font-bold tracking-tight transition-colors duration-500 ${isRunning ? "text-white" : "text-gray-200"}`}>{isRunning ? "Connected" : "Disconnected"}</h2><p className="text-xs text-gray-500 font-mono mt-1 mb-6">{status}</p><div className={`flex justify-center gap-3 max-w-[320px] mx-auto transition-all duration-500 ${isRunning ? "opacity-100 translate-y-0" : "opacity-0 translate-y-4"}`}><div className="bg-black/20 rounded-xl p-3 border border-white/5 flex-1 flex flex-col items-center justify-center min-w-[120px]"><div className="text-[9px] text-gray-500 uppercase mb-1 flex items-center gap-1.5 font-bold tracking-wider"><svg className="w-3 h-3 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="M19 14l-7 7m0 0l-7-7m7 7V3" /></svg>DOWN</div><div className="text-sm font-mono font-bold text-white whitespace-nowrap">{formatSpeed(traffic.down)}</div></div><div className="bg-black/20 rounded-xl p-3 border border-white/5 flex-1 flex flex-col items-center justify-center min-w-[120px]"><div className="text-[9px] text-gray-500 uppercase mb-1 flex items-center gap-1.5 font-bold tracking-wider"><svg className="w-3 h-3 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="M5 10l7-7m0 0l7 7m-7-7v18" /></svg>UP</div><div className="text-sm font-mono font-bold text-white whitespace-nowrap">{formatSpeed(traffic.up)}</div></div></div></div>
                    </div>
                </div>

                <div className="glass w-80 rounded-3xl p-6 flex flex-col relative">
                    <div className="flex justify-between items-center pb-4 border-b border-white/5 mb-5">
                        <span className="text-gray-400 text-[10px] font-bold tracking-widest uppercase">Servers</span>
                        <div className="flex items-center gap-2">
                            <button onClick={onPing} className={`w-6 h-6 rounded flex items-center justify-center transition-all bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white`}><svg xmlns="http://www.w3.org/2000/svg" className={`h-3.5 w-3.5 ${isPinging ? "animate-spin text-purple-400" : ""}`} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg></button>
                            <button onClick={() => setIsAdding(!isAdding)} className={`w-6 h-6 rounded flex items-center justify-center transition-all duration-500 ${isAdding ? "bg-purple-600 text-white rotate-45" : "bg-white/5 hover:bg-white/10 text-gray-400"}`}><svg xmlns="http://www.w3.org/2000/svg" className="h-3 w-3 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" /></svg></button>
                        </div>
                    </div>

                    <div
                        className={`grid transition-[grid-template-rows] duration-500 ease-in-out ${isAdding ? "grid-rows-[1fr]" : "grid-rows-[0fr]"}`}
                        style={{ willChange: "grid-template-rows" }}
                    >
                        <div className="overflow-hidden">
                            <div className={`pb-4 transition-all duration-500 ${isAdding ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-2"}`}>
                                <div className="bg-black/20 p-3 rounded-xl border border-white/5">
                                    <div className="flex gap-2 mb-2">
                                        <button onClick={() => setAddType("key")} className={`flex-1 text-[10px] py-1 rounded transition-colors ${addType === "key" ? "bg-white/10 text-white" : "text-gray-500 hover:text-gray-300"}`}>KEY</button>
                                        <button onClick={() => setAddType("sub")} className={`flex-1 text-[10px] py-1 rounded transition-colors ${addType === "sub" ? "bg-white/10 text-white" : "text-gray-500 hover:text-gray-300"}`}>SUB</button>
                                    </div>
                                    <input
                                        ref={inputRef}
                                        type="text"
                                        placeholder={addType === "key" ? "vless://..." : "https://..."}
                                        value={inputVal}
                                        onChange={(e) => setInputVal(e.target.value)}
                                        onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
                                        className="w-full bg-transparent border-b border-white/10 px-1 py-2 text-xs text-white outline-none focus:border-purple-500 mb-3 placeholder:text-gray-700 font-mono"
                                    />
                                    <button onClick={handleAdd} disabled={isProcessing} className="w-full bg-purple-600 hover:bg-purple-500 disabled:opacity-50 text-white text-[10px] font-bold py-2 rounded-lg transition-colors active:scale-95">
                                        {isProcessing ? "PROCESSING..." : "ADD"}
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div className="flex-1 overflow-y-auto pr-1 scrollbar-hide">
                        {manualProfiles.length > 0 && ( <div className="mb-4 animate-slideIn"><div className="text-[9px] font-bold text-gray-600 uppercase mb-2 px-1">Manual</div>{manualProfiles.map(renderProfileItem)}</div> )}
                        {subscriptions.map(sub => (
                            <div key={sub.id} className="mb-4 animate-slideIn">
                                <div className="flex justify-between items-center mb-2 px-1 group">
                                    <div className="text-[9px] font-bold text-purple-400/70 uppercase truncate max-w-[120px]">{sub.name}</div>
                                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <button onClick={(e) => handleUpdateSub(e, sub.id)} className="text-gray-500 hover:text-white p-1" title="Update"><svg className={`w-3 h-3 ${updatingSubId === sub.id ? "animate-spin" : ""}`} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg></button>
                                        <button onClick={(e) => requestDeleteSub(e, sub.id)} className="text-gray-500 hover:text-red-400 p-1" title="Delete"><svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
                                    </div>
                                </div>
                                {getSubProfiles(sub.id).map(renderProfileItem)}
                            </div>
                        ))}
                        {profiles.length === 0 && ( <div className="h-full flex flex-col items-center justify-center text-gray-700 text-xs text-center border-2 border-dashed border-white/5 rounded-xl"><span>No servers</span></div> )}
                    </div>
                </div>
            </div>

            <ConfirmationModal
                isOpen={!!subToDelete}
                title="Delete Subscription"
                message="Are you sure you want to delete this subscription? All associated profiles will be removed."
                onClose={() => setSubToDelete(null)}
                onConfirm={confirmDeleteSub}
            />

            <EditProfileModal
                isOpen={!!profileToEdit}
                initialName={profileToEdit?.name || ""}
                initialKey={profileToEdit?.key || ""}
                onClose={() => setProfileToEdit(null)}
                onSave={handleSaveProfile}
            />
        </>
    );
};