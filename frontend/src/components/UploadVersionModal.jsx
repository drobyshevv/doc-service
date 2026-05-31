// frontend/src/components/UploadVersionModal.jsx
import { useState } from 'react';
import Modal from './Modal';
import { documentsAPI } from '../services/api';

export default function UploadVersionModal({ isOpen, onClose, documentId, onSuccess }) {
  const [file, setFile] = useState(null);
  const [note, setNote] = useState('');
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!file) {
      setError('Выберите файл');
      return;
    }
    
    setUploading(true);
    setError('');
    
    try {
      const formData = new FormData();
      formData.append('file', file);
      if (note.trim()) {
        formData.append('note', note.trim());
      }
      
      await documentsAPI.uploadVersion(documentId, formData);
      
      // Успех: сброс формы + коллбек
      setFile(null);
      setNote('');
      onSuccess?.();
      onClose();
      
    } catch (err) {
      console.error('Upload version error:', err);
      setError(err.message || 'Не удалось загрузить версию');
    } finally {
      setUploading(false);
    }
  };

  const handleFileChange = (e) => {
    const selectedFile = e.target.files?.[0];
    if (selectedFile) {
      setFile(selectedFile);
      setError('');
    }
  };

  const handleClose = () => {
    setFile(null);
    setNote('');
    setError('');
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="🔄 Загрузить новую версию">
      <form onSubmit={handleSubmit} className="modal-form">
        
        {/* Область выбора файла */}
        <label className={`modal-form__file ${file ? 'modal-form__file--active' : ''}`}>
          <input
            type="file"
            className="modal-form__file-input"
            onChange={handleFileChange}
            disabled={uploading}
            accept="*/*"
          />
          <div className="modal-form__file-label">
            <span className="modal-form__file-icon">📁</span>
            <span>{file ? file.name : 'Выберите файл или перетащите сюда'}</span>
          </div>
          {file && (
            <div className="modal-form__file-name">
              {file.name} • {(file.size / 1024).toFixed(1)} KB
            </div>
          )}
        </label>

        {/* Комментарий к версии */}
        <div className="form-group">
          <label htmlFor="version-note" className="form-label">
            Комментарий (опционально)
          </label>
          <textarea
            id="version-note"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            className="input"
            placeholder="Что изменилось в этой версии?"
            rows={3}
            disabled={uploading}
          />
        </div>

        {/* Ошибка */}
        {error && (
          <div className="page-auth__error" role="alert">
            {error}
          </div>
        )}

        {/* Кнопки */}
        <div className="modal-form__actions">
          <button
            type="button"
            onClick={handleClose}
            className="btn btn-outline"
            disabled={uploading}
          >
            Отмена
          </button>
          <button
            type="submit"
            className="btn btn-primary"
            disabled={uploading || !file}
          >
            {uploading ? 'Загрузка...' : 'Загрузить версию'}
          </button>
        </div>
        
      </form>
    </Modal>
  );
}