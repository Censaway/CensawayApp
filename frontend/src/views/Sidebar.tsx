import React from 'react';

interface Props { view: string; setView: (v: any) => void; }

export const Sidebar: React.FC<Props> = ({ view, setView }) => {
    return (
        <div className="w-20 glass h-full border-r border-white/5 flex flex-col items-center py-6 z-50 shrink-0" style={{ "--wails-draggable": "drag" } as React.CSSProperties}>
            <div className="mb-10 w-8 h-8 rounded-full bg-gradient-to-tr from-purple-500 to-pink-500 shadow-[0_0_15px_rgba(168,85,247,0.5)]"></div>
            <div className="flex flex-col gap-6 w-full px-2" style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}>
                <NavBtn icon={<svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>} label="HOME" active={view === "dashboard"} onClick={() => setView("dashboard")} />
                <NavBtn icon={<svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>} label="LOGS" active={view === "logs"} onClick={() => setView("logs")} />
                <NavBtn icon={<svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7" /></svg>} label="ROUTING" active={view === "routing"} onClick={() => setView("routing")} />
                <NavBtn icon={<svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>} label="SETTINGS" active={view === "settings"} onClick={() => setView("settings")} />
            </div>
        </div>
    );
}

const NavBtn = ({ icon, label, active, onClick }: any) => (
    <button onClick={onClick} className={`group flex flex-col items-center gap-1 p-2 rounded-xl transition-all ${active ? "bg-white/10 text-white shadow-lg shadow-purple-500/10" : "text-gray-500 hover:text-gray-300 hover:bg-white/5"}`}>
        {icon}
        <span className="text-[9px] font-bold tracking-wider opacity-0 group-hover:opacity-100 transition-opacity absolute left-16 bg-black/80 px-2 py-1 rounded border border-white/10 pointer-events-none z-50">{label}</span>
    </button>
);