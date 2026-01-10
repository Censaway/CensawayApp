import React from 'react';
import { main } from "../../wailsjs/go/models";
import { RestartBanner } from '../components/RestartBanner';

interface Props {
    settings: main.Settings;
    onUpdate: (s: main.Settings) => void;
    hasChanges: boolean;
    isRunning: boolean;
    onRestart: () => void;
}

export const SettingsView: React.FC<Props> = ({ settings, onUpdate, hasChanges, isRunning, onRestart }) => {

    const update = (changes: Partial<main.Settings>) => {
        const newS = new main.Settings({ ...settings, ...changes });
        onUpdate(newS);
    };

    const isProxy = settings.run_mode === "proxy";

    return (
        <div className="w-full max-w-2xl animate-[fadeIn_0.3s_ease-out]">
            <div className="glass rounded-3xl p-8 border-t border-white/10">
                <h2 className="text-xl font-bold mb-6 text-gray-200 tracking-tight">Configuration</h2>

                
                <div className="mb-8">
                    <div className="text-sm font-medium text-white mb-3">Split Tunneling</div>
                    <div className="grid grid-cols-2 gap-3">
                        <button onClick={() => update({ routing_mode: "smart" })} className={`p-4 rounded-xl border text-left transition-all ${settings.routing_mode === "smart" ? "bg-purple-500/20 border-purple-500/50 shadow-[0_0_15px_rgba(168,85,247,0.15)]" : "bg-black/20 border-white/5 hover:bg-white/5 opacity-70 hover:opacity-100"}`}><div className={`font-bold text-sm mb-1 ${settings.routing_mode === "smart" ? "text-purple-300" : "text-gray-400"}`}>Smart Mode</div><div className="text-[10px] text-gray-500 leading-tight">Bypass local sites. Proxy blocked only.</div></button>
                        <button onClick={() => update({ routing_mode: "global" })} className={`p-4 rounded-xl border text-left transition-all ${settings.routing_mode === "global" ? "bg-purple-500/20 border-purple-500/50 shadow-[0_0_15px_rgba(168,85,247,0.15)]" : "bg-black/20 border-white/5 hover:bg-white/5 opacity-70 hover:opacity-100"}`}><div className={`font-bold text-sm mb-1 ${settings.routing_mode === "global" ? "text-purple-300" : "text-gray-400"}`}>Global Mode</div><div className="text-[10px] text-gray-500 leading-tight">Route ALL traffic through VPN.</div></button>
                    </div>
                </div>

                
                <div className="mb-8">
                    <div className="text-sm font-medium text-white mb-3">Operating Mode</div>
                    <div className="grid grid-cols-2 gap-3 mb-3">
                        <button onClick={() => update({ run_mode: "tun" })} className={`p-4 rounded-xl border text-left transition-all ${settings.run_mode === "tun" ? "bg-emerald-500/20 border-emerald-500/50 shadow-[0_0_15px_rgba(16,185,129,0.15)]" : "bg-black/20 border-white/5 hover:bg-white/5 opacity-70 hover:opacity-100"}`}><div className={`font-bold text-sm mb-1 ${settings.run_mode === "tun" ? "text-emerald-300" : "text-gray-400"}`}>TUN Mode</div><div className="text-[10px] text-gray-500 leading-tight">Virtual Interface. All apps.</div></button>
                        <button onClick={() => update({ run_mode: "proxy" })} className={`p-4 rounded-xl border text-left transition-all ${settings.run_mode === "proxy" ? "bg-emerald-500/20 border-emerald-500/50 shadow-[0_0_15px_rgba(16,185,129,0.15)]" : "bg-black/20 border-white/5 hover:bg-white/5 opacity-70 hover:opacity-100"}`}><div className={`font-bold text-sm mb-1 ${settings.run_mode === "proxy" ? "text-emerald-300" : "text-gray-400"}`}>System Proxy</div><div className="text-[10px] text-gray-500 leading-tight">Browsers only.</div></button>
                    </div>

                    
                    <div className={`grid transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)] ${isProxy ? "grid-rows-[1fr] opacity-100 mt-4" : "grid-rows-[0fr] opacity-0 mt-0"}`}>
                        <div className="overflow-hidden min-h-0">
                            <div className="flex items-center justify-between bg-white/5 p-4 rounded-xl border border-white/5">
                                <div className="flex flex-col"><span className="text-sm font-medium text-gray-200">Listening Port</span><span className="text-[10px] text-gray-500">Local SOCKS/HTTP port</span></div>
                                <div className="relative"><div className="absolute inset-y-0 left-3 flex items-center pointer-events-none"><span className="text-emerald-500/50 text-xs font-mono">:</span></div><input type="number" value={settings.mixed_port} onChange={(e) => update({ mixed_port: parseInt(e.target.value) || 0 })} className="w-24 bg-black/40 border border-white/10 rounded-lg py-2 pl-6 pr-3 text-right text-sm text-emerald-400 font-mono outline-none focus:border-emerald-500/50 focus:bg-black/60 transition-all [&::-webkit-inner-spin-button]:appearance-none" placeholder="2080" /></div>
                            </div>
                        </div>
                    </div>
                </div>

                
                <div className="mb-8">
                    <div className="text-sm font-medium text-white mb-3">Startup</div>
                    <div
                        onClick={() => update({ auto_connect: !settings.auto_connect })}
                        className={`group flex items-center justify-between p-4 rounded-xl border cursor-pointer transition-all ${settings.auto_connect ? "bg-purple-500/10 border-purple-500/30 shadow-[0_0_20px_-5px_rgba(168,85,247,0.2)]" : "bg-black/20 border-white/5 hover:bg-white/5"}`}
                    >
                        <div className="flex flex-col">
                            <span className={`text-sm font-bold transition-colors ${settings.auto_connect ? "text-purple-400" : "text-gray-400"}`}>Auto Connect</span>
                            <span className="text-[10px] text-gray-500">Connect to last server on launch</span>
                        </div>
                        
                        <div className={`w-10 h-5 rounded-full relative transition-colors ${settings.auto_connect ? "bg-purple-600" : "bg-white/10"}`}>
                            <div className={`absolute top-1 left-1 w-3 h-3 rounded-full bg-white shadow-sm transition-transform ${settings.auto_connect ? "translate-x-5" : "translate-x-0"}`}></div>
                        </div>
                    </div>
                </div>

                <RestartBanner visible={isRunning && hasChanges} onRestart={onRestart} />
            </div>
        </div>
    );
};