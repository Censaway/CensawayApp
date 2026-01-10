import React from 'react';

interface Props {
    visible: boolean;
    onRestart: () => void;
}

export const RestartBanner: React.FC<Props> = ({ visible, onRestart }) => {
    return (
        <div
            className={`
                grid transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)]
                ${visible ? "grid-rows-[1fr] opacity-100 mt-6" : "grid-rows-[0fr] opacity-0 mt-0"}
            `}
        >
            <div className="overflow-hidden min-h-0">
                <button
                    onClick={onRestart}
                    className="w-full bg-yellow-500/10 hover:bg-yellow-500/20 text-yellow-500 text-xs font-bold py-3 rounded-xl border border-yellow-500/20 transition-all shadow-[0_0_15px_-5px_rgba(234,179,8,0.3)] hover:shadow-[0_0_20px_-5px_rgba(234,179,8,0.5)]"
                >
                    ⚠️ SETTINGS CHANGED - RESTART CORE TO APPLY
                </button>
            </div>
        </div>
    );
};