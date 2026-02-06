import React, { useState, useEffect } from 'react';
import { GetMixedProfiles, CreateMixedProfile, DeleteMixedProfile, UpdateMixedProfile, StartVless, StopVless } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";
import { useAppStore } from '../store/appStore';
import { CustomSelect } from '../components/CustomSelect';

export const MixerView: React.FC = () => {
    const { profiles, connectionState, setConnectionState, setStatus, setErrorMsg, setCurrentIp, setTraffic, selectedId, setSelectedId } = useAppStore();
    const [mixedProfiles, setMixedProfiles] = useState<main.MixedProfile[]>([]);
    const [isEditorOpen, setIsEditorOpen] = useState(false);
    const [editId, setEditId] = useState<string | null>(null);
    const [editName, setEditName] = useState("");
    const [relayId, setRelayId] = useState("");
    const [exitId, setExitId] = useState("");

    const isRunning = connectionState === "connected";

    useEffect(() => {
        loadMixed();
    }, []);

    const loadMixed = async () => {
        try {
            const list = await GetMixedProfiles();
            setMixedProfiles(list || []);
        } catch(e) { console.error(e); }
    };

    const handleCreate = () => {
        setEditId(null);
        setEditName("New Chain");
        setRelayId("");
        setExitId("");
        setIsEditorOpen(true);
    };

    const handleEdit = (m: main.MixedProfile) => {
        setEditId(m.id);
        setEditName(m.name);
        setRelayId(m.relay_id);
        setExitId(m.exit_id);
        setIsEditorOpen(true);
    };

    const handleSave = async () => {
        if (!editName) return;
        if (!relayId || !exitId) {
             alert("Select both relay and exit nodes");
             return;
        }
        if (relayId === exitId) {
            alert("Relay and Exit nodes must be different");
            return;
        }

        if (editId) {
            await UpdateMixedProfile(editId, editName, relayId, exitId);
        } else {
            await CreateMixedProfile(editName, relayId, exitId);
        }
        setIsEditorOpen(false);
        loadMixed();
    };

    const handleDelete = async (e: React.MouseEvent, id: string) => {
        e.stopPropagation();
        if(confirm("Delete this chain?")) {
            await DeleteMixedProfile(id);
            loadMixed();
        }
    };

    const handleRun = async (m: main.MixedProfile) => {
        const mixerId = `mixed://${m.id}`;
        
        if (isRunning) {
            if (selectedId === mixerId) {
                 setStatus("Stopping..."); await StopVless();
                 setStatus("Disconnected"); setConnectionState("disconnected"); setTraffic({ up: 0, down: 0 }); setCurrentIp(null);
            } else {
                 setStatus("Switching..."); await StopVless();
                 startMixer(mixerId);
            }
        } else {
            startMixer(mixerId);
        }
    };

    const startMixer = async (id: string) => {
         setSelectedId(id);
         setConnectionState("connecting"); setStatus("Starting Chain...");
         const res = await StartVless(id);
         if (res === "Connected") {
             setConnectionState("connected"); setStatus("Secured (Chain)"); setCurrentIp(null);
         } else {
             setConnectionState("disconnected"); setStatus(res); setErrorMsg(res);
         }
    };

    const profileOptions = profiles.map(p => ({
        value: p.id,
        label: p.name
    }));

    return (
        <div className="w-full max-w-5xl h-[520px] flex gap-6 animate-[fadeIn_0.3s_ease-out]">
            <div className="glass flex-1 rounded-3xl p-8 border-t border-white/10 flex flex-col">
                 <div className="flex justify-between items-center mb-6">
                    <div>
                        <h2 className="text-xl font-bold text-white tracking-tight">Chain Mixer</h2>
                        <p className="text-[10px] text-gray-500 mt-1">Route: You &rarr; Relay &rarr; Exit &rarr; Internet</p>
                    </div>
                    <button onClick={handleCreate} className="bg-purple-600 hover:bg-purple-500 text-white px-4 py-2 rounded-xl text-xs font-bold shadow-lg shadow-purple-500/20 transition-all active:scale-95">CREATE CHAIN</button>
                 </div>

                 <div className="flex-1 overflow-y-auto pr-1 scrollbar-hide space-y-3">
                     {mixedProfiles.length === 0 && (
                        <div className="h-40 flex flex-col items-center justify-center text-gray-600 text-xs border-2 border-dashed border-white/5 rounded-xl">
                            No chains created
                        </div>
                     )}
                     {mixedProfiles.map(m => {
                         const isActive = isRunning && selectedId === `mixed://${m.id}`;
                         const relayName = profiles.find(p => p.id === m.relay_id)?.name || "Unknown";
                         const exitName = profiles.find(p => p.id === m.exit_id)?.name || "Unknown";

                         return (
                            <div key={m.id} className={`p-4 rounded-xl border transition-all flex justify-between items-center group ${isActive ? "bg-purple-500/10 border-purple-500/30" : "bg-white/5 border-transparent hover:bg-white/10"}`}>
                                <div>
                                    <div className="flex items-center gap-2">
                                        <h3 className={`font-bold text-sm ${isActive ? "text-white" : "text-gray-200"}`}>{m.name}</h3>
                                        <span className="text-[9px] px-1.5 py-0.5 rounded uppercase font-mono border text-purple-400 border-purple-500/20 bg-purple-500/10">CHAIN</span>
                                    </div>
                                    <div className="text-[10px] text-gray-500 mt-1.5 flex items-center gap-2">
                                        <span className="max-w-[80px] truncate">{relayName}</span>
                                        <svg className="w-3 h-3 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" /></svg>
                                        <span className="max-w-[80px] truncate text-gray-300">{exitName}</span>
                                    </div>
                                </div>
                                
                                <div className="flex items-center gap-3">
                                    <button onClick={() => handleRun(m)} className={`w-10 h-10 rounded-full flex items-center justify-center transition-all ${isActive ? "bg-red-500/10 text-red-400 hover:bg-red-500/20" : "bg-green-500/10 text-green-400 hover:bg-green-500/20"}`}>
                                        {isActive ? (
                                            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                                        ) : (
                                            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                                        )}
                                    </button>
                                    <button onClick={() => handleEdit(m)} className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white transition-colors"><svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg></button>
                                    <button onClick={(e) => handleDelete(e, m.id)} className="p-2 rounded-lg bg-white/5 hover:bg-red-500/10 text-gray-400 hover:text-red-400 transition-colors"><svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
                                </div>
                            </div>
                         );
                     })}
                 </div>
            </div>

            <div className={`glass w-80 rounded-3xl p-6 border-t border-white/10 flex flex-col transition-all duration-300 ${isEditorOpen ? "translate-x-0 opacity-100" : "translate-x-10 opacity-50 pointer-events-none grayscale"}`}>
                <div className="mb-6">
                    <h3 className="text-sm font-bold text-white">{editId ? "Edit Chain" : "New Chain"}</h3>
                </div>
                
                <div className="space-y-6 flex-1">
                    <div>
                        <label className="text-[9px] font-bold text-gray-500 uppercase tracking-wider mb-1.5 ml-1 block">Name</label>
                        <input type="text" value={editName} onChange={(e) => setEditName(e.target.value)} className="w-full h-9 bg-[#0a0a0e] border border-white/10 rounded-lg px-3 text-xs text-white outline-none focus:border-purple-500" />
                    </div>

                    <div className="relative">
                        <div className="absolute left-3 top-8 bottom-0 w-0.5 bg-gradient-to-b from-purple-500/50 to-transparent -z-10"></div>
                        
                        <div className="mb-4">
                            <label className="text-[9px] font-bold text-purple-400 uppercase tracking-wider mb-1.5 ml-1 flex items-center gap-1">
                                <span className="w-1.5 h-1.5 rounded-full bg-purple-500"></span>
                                Relay Node (Entry)
                            </label>
                            <CustomSelect value={relayId} onChange={setRelayId} options={profileOptions} className="w-full" />
                            <p className="text-[9px] text-gray-600 mt-1 ml-1">Traffic connects here first</p>
                        </div>

                        <div>
                            <label className="text-[9px] font-bold text-white uppercase tracking-wider mb-1.5 ml-1 flex items-center gap-1">
                                <span className="w-1.5 h-1.5 rounded-full bg-white"></span>
                                Exit Node (Final)
                            </label>
                            <CustomSelect value={exitId} onChange={setExitId} options={profileOptions} className="w-full" />
                            <p className="text-[9px] text-gray-600 mt-1 ml-1">Traffic exits to internet here</p>
                        </div>
                    </div>
                </div>

                <div className="mt-4 flex gap-3">
                    <button onClick={() => setIsEditorOpen(false)} className="flex-1 py-2 rounded-xl text-[10px] font-bold bg-white/5 hover:bg-white/10 text-gray-400">CANCEL</button>
                    <button onClick={handleSave} className="flex-1 py-2 rounded-xl text-[10px] font-bold bg-purple-600 hover:bg-purple-500 text-white shadow-lg shadow-purple-500/20">SAVE CHAIN</button>
                </div>
            </div>
        </div>
    );
};