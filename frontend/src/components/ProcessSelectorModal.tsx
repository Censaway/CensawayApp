import React, { useState, useEffect, useMemo } from 'react';
import { GetRunningProcesses } from '../../wailsjs/go/main/App';

interface Props {
    isOpen: boolean;
    onClose: () => void;
    onSelect: (processName: string) => void;
}

export const ProcessSelectorModal: React.FC<Props> = ({ isOpen, onClose, onSelect }) => {
    const [processes, setProcesses] = useState<string[]>([]);
    const [search, setSearch] = useState("");
    const [loading, setLoading] = useState(false);
    const [isVisible, setIsVisible] = useState(false);
    const [shouldRender, setShouldRender] = useState(false);

    useEffect(() => {
        if (isOpen) {
            setShouldRender(true);
            setTimeout(() => setIsVisible(true), 50);
            fetchProcesses();
        } else {
            setIsVisible(false);
            const timer = setTimeout(() => setShouldRender(false), 300);
            return () => clearTimeout(timer);
        }
    }, [isOpen]);

    const fetchProcesses = async () => {
        setLoading(true);
        try {
            const list = await GetRunningProcesses();
            setProcesses(list);
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    const filtered = useMemo(() => {
        return processes.filter(p => p.toLowerCase().includes(search.toLowerCase()));
    }, [processes, search]);

    if (!shouldRender) return null;

    return (
        <div 
            className={`fixed inset-0 z-[200] flex items-center justify-center bg-black/80 backdrop-blur-md transition-opacity duration-300 ${isVisible ? "opacity-100" : "opacity-0"}`}
            onClick={onClose}
        >
            <div 
                className={`w-[400px] bg-[#18181b] p-6 rounded-2xl border border-white/10 shadow-2xl flex flex-col max-h-[80vh] transform transition-all duration-300 ${isVisible ? "scale-100 translate-y-0" : "scale-95 translate-y-4"}`}
                onClick={e => e.stopPropagation()}
            >
                <div className="flex justify-between items-center mb-4">
                    <h3 className="text-sm font-bold text-white flex items-center gap-2">
                        <svg className="w-4 h-4 text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" /></svg>
                        Select Process
                    </h3>
                    <button onClick={fetchProcesses} className="p-1.5 rounded-lg bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white transition-colors" title="Refresh">
                        <svg className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
                    </button>
                </div>

                <input 
                    type="text" 
                    placeholder="Search process..." 
                    value={search}
                    onChange={e => setSearch(e.target.value)}
                    autoFocus
                    className="w-full bg-[#0a0a0e] border border-white/10 rounded-xl px-3 py-2.5 text-xs text-white outline-none focus:border-purple-500 mb-4 transition-colors"
                />

                <div className="flex-1 overflow-y-auto min-h-0 pr-1 scrollbar-thin space-y-1">
                    {loading && processes.length === 0 ? (
                        <div className="flex justify-center py-8 text-gray-500 text-xs">Loading...</div>
                    ) : filtered.length === 0 ? (
                        <div className="flex justify-center py-8 text-gray-500 text-xs">No processes found</div>
                    ) : (
                        filtered.map((proc, idx) => (
                            <button
                                key={`${proc}-${idx}`}
                                onClick={() => { onSelect(proc); onClose(); }}
                                className="w-full text-left px-3 py-2 rounded-lg bg-white/5 hover:bg-white/10 text-xs text-gray-300 hover:text-white transition-colors flex items-center gap-2 group"
                            >
                                <span className="w-1.5 h-1.5 rounded-full bg-green-500/50 group-hover:bg-green-400"></span>
                                {proc}
                            </button>
                        ))
                    )}
                </div>

                <div className="mt-4 pt-4 border-t border-white/5 flex justify-end">
                    <button onClick={onClose} className="px-4 py-2 rounded-lg text-[10px] font-bold text-gray-400 hover:text-white bg-white/5 hover:bg-white/10 transition-colors">
                        CANCEL
                    </button>
                </div>
            </div>
        </div>
    );
};