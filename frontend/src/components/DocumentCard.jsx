// frontend/src/components/DocumentCard.jsx
import { Link } from 'react-router-dom';
import { documentsAPI } from '../services/api';

export default function DocumentCard({ doc, showActions = false, onDownload }) {
  const id = doc.ID || doc.id;
  const title = doc.Title || doc.title || doc.OriginalFilename || 'Без названия';
  const version = doc.CurrentVersion || doc.current_version || 1;
  const isPublic = doc.IsPublic || doc.is_public;
  const mimeType = doc.MimeType || doc.mime_type || 'application/octet-stream';

  // Иконка по типу файла
  const getFileIcon = (mime) => {
    if (mime.startsWith('image/')) return '🖼️';
    if (mime.startsWith('video/')) return '🎬';
    if (mime.startsWith('audio/')) return '🎵';
    if (mime.includes('pdf')) return '📕';
    if (mime.includes('word') || mime.includes('document')) return '📘';
    if (mime.includes('excel') || mime.includes('spreadsheet')) return '📗';
    if (mime.includes('powerpoint') || mime.includes('presentation')) return '📙';
    if (mime.includes('text') || mime.includes('markdown')) return '📝';
    return '📄';
  };

  const handleDownload = (e) => {
    e.preventDefault();
    e.stopPropagation();
    documentsAPI.download(id).then(r => {
      const a = document.createElement('a');
      a.href = URL.createObjectURL(r.data);
      a.download = title;
      a.click();
      URL.revokeObjectURL(a.href);
      onDownload?.();
    });
  };

  return (
    <Link 
      to={`/document/${id}`}
      className="card block group hover:border-indigo-300 transition-all"
    >
      <div className="flex items-start justify-between gap-3">
        {/* Иконка + название */}
        <div className="flex items-start gap-3 min-w-0 flex-1">
          <div className="w-10 h-10 rounded-lg bg-slate-100 flex items-center justify-center text-xl shrink-0">
            {getFileIcon(mimeType)}
          </div>
          <div className="min-w-0">
            <h3 className="font-medium text-slate-800 truncate group-hover:text-indigo-600 transition-colors">
              {title}
            </h3>
            <div className="flex items-center gap-2 mt-1">
              <span className="badge badge-primary">v{version}</span>
              {isPublic && <span className="badge badge-success">Публичный</span>}
            </div>
          </div>
        </div>

        {/* Кнопки действий (показываются при ховере или на мобильных) */}
        {showActions && (
          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <button 
              onClick={handleDownload}
              className="btn btn-ghost btn-icon text-slate-500 hover:text-indigo-600"
              title="Скачать"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </button>
            <button className="btn btn-ghost btn-icon text-slate-500 hover:text-slate-700" title="Ещё">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
              </svg>
            </button>
          </div>
        )}
      </div>
    </Link>
  );
}