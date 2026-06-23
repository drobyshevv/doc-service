// frontend/src/components/DocumentCard.jsx
import { Link } from 'react-router-dom';
import { documentsAPI } from '../services/api';

export default function DocumentCard({ doc, showActions = false, onDownload }) {
  const id = doc.ID || doc.id;
  const title = doc.Title || doc.title || doc.OriginalFilename || 'Без названия';
  const version = doc.CurrentVersion || doc.current_version || 1;
  const isPublic = doc.IsPublic || doc.is_public;
  const mimeType = doc.MimeType || doc.mime_type || 'application/octet-stream';
  const fileSize = doc.FileSize || doc.file_size || 0;

  const getExtension = (mime, filename) => {
    if (filename) {
      const ext = filename.split('.').pop().toLowerCase();
      return ext.toUpperCase();
    }
    if (mime.includes('pdf')) return 'PDF';
    if (mime.includes('word') || mime.includes('document')) return 'DOCX';
    if (mime.includes('text') || mime.includes('markdown')) return 'TXT';
    if (mime.includes('excel') || mime.includes('spreadsheet')) return 'XLSX';
    if (mime.includes('powerpoint') || mime.includes('presentation')) return 'PPTX';
    if (mime.includes('image')) return 'IMG';
    return 'FILE';
  };

  const extension = getExtension(mimeType, doc.OriginalFilename || title);

  const formatFileSize = (bytes) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
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
    <Link to={`/document/${id}`} className="file-card">
      <div className="file-card__paper">
        <div className="file-card__fold"></div>
        <div className="file-card__content">
          <span className="file-card__extension">{extension}</span>
        </div>
      </div>
      <div className="file-card__info">
        <h3 className="file-card__title">{title}</h3>
        <div className="file-card__meta">
          <span className="badge badge-primary">v{version}</span>
          {isPublic && <span className="badge badge-success">Публичный</span>}
          <span className="file-card__size">{formatFileSize(fileSize)}</span>
        </div>
      </div>
      {showActions && (
        <button 
          onClick={handleDownload}
          className="file-card__download"
          title="Скачать"
        >
          <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
        </button>
      )}
    </Link>
  );
}