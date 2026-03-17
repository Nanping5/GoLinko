import React from 'react';
import { AlertTriangle } from 'lucide-react';

export interface ConfirmDialogProps {
  title: string;
  message: string;
  confirmLabel?: string;
  danger?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  title,
  message,
  confirmLabel = '确认',
  danger = false,
  onConfirm,
  onCancel,
}) => (
  <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[300] flex items-center justify-center p-4">
    <div className="bg-white rounded-2xl w-full max-w-xs shadow-2xl animate-in zoom-in duration-150">
      <div className="p-5">
        <div className="flex items-start gap-3">
          {danger && (
            <div className="w-9 h-9 rounded-xl bg-red-100 flex items-center justify-center shrink-0">
              <AlertTriangle className="w-5 h-5 text-red-500" />
            </div>
          )}
          <div>
            <h3 className="font-bold text-gray-900 text-base leading-tight">{title}</h3>
            <p className="text-sm text-gray-500 mt-1.5 leading-relaxed">{message}</p>
          </div>
        </div>
      </div>
      <div className="flex gap-2 px-5 pb-5">
        <button
          onClick={onCancel}
          className="flex-1 py-2.5 rounded-xl bg-gray-100 text-gray-700 text-sm font-medium hover:bg-gray-200 active:scale-95 transition-all"
        >
          取消
        </button>
        <button
          onClick={onConfirm}
          className={`flex-1 py-2.5 rounded-xl text-sm font-medium text-white active:scale-95 transition-all
            ${danger ? 'bg-red-500 hover:bg-red-600' : 'bg-[#07c160] hover:bg-[#06ad55]'}`}
        >
          {confirmLabel}
        </button>
      </div>
    </div>
  </div>
);
