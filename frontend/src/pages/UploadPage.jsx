// frontend/src/pages/UploadPage.jsx
import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { documentsAPI } from '../services/api';

export default function UploadPage() {
  const [file, setFile] = useState(null);
  const [title, setTitle] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!file) {
      setError('Выберите файл для загрузки');
      return;
    }
    
    setUploading(true);
    setError('');
    
    try {
      const formData = new FormData();
      formData.append('file', file);
      if (title) formData.append('title', title);
      formData.append('is_public', isPublic ? 'true' : 'false');
      
      await documentsAPI.upload(formData);
      navigate('/');
    } catch (err) {
      console.error('Upload error:', err);
      setError(err.response?.data || err.message || 'Ошибка загрузки документа');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="page-auth">
      <div className="page-auth__card">
        {/* Логотип */}
        <div className="page-auth__logo">
          <Link to="/" className="logo-link">
            <div className="logo">D</div>
            <span>DocService</span>
          </Link>
        </div>

        {/* Заголовок */}
        <div className="page-auth__header">
          <h1 className="page-auth__title">Загрузка документа</h1>
          <p className="page-auth__subtitle">
            Добавьте новый документ в систему для индексации и поиска
          </p>
        </div>

        {/* Ошибка */}
        {error && (
          <div className="page-auth__error" role="alert">
            {error}
          </div>
        )}

        {/* Форма */}
        <form onSubmit={handleSubmit} className="page-auth__form">
          <div className="form-group">
            <label htmlFor="title" className="form-label">
              Название документа
            </label>
            <input
              id="title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="input"
              placeholder="По умолчанию — имя файла"
              disabled={uploading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="file" className="form-label">
              Файл <span style={{ color: 'var(--danger, #e74c3c)' }}>*</span>
            </label>
            <input
              id="file"
              type="file"
              onChange={(e) => setFile(e.target.files[0])}
              className="input"
              required
              disabled={uploading}
              accept=".pdf,.docx,.txt,.doc"
            />
            {file && (
              <p style={{ 
                marginTop: '0.5rem', 
                fontSize: '0.875rem', 
                color: 'var(--text-muted, #7f8c8d)' 
              }}>
                Выбран: <strong>{file.name}</strong> ({(file.size / 1024).toFixed(1)} КБ)
              </p>
            )}
          </div>

          <div className="form-group">
            <label 
              htmlFor="isPublic" 
              className="form-label"
              style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}
            >
              <input
                id="isPublic"
                type="checkbox"
                checked={isPublic}
                onChange={(e) => setIsPublic(e.target.checked)}
                disabled={uploading}
                style={{ width: '18px', height: '18px', cursor: 'pointer' }}
              />
              <span>🌐 Публичный документ (виден всем пользователям)</span>
            </label>
            <p style={{ 
              marginTop: '0.25rem', 
              fontSize: '0.8125rem', 
              color: 'var(--text-muted, #7f8c8d)',
              marginLeft: '1.75rem'
            }}>
              Приватные документы видны только вам
            </p>
          </div>

          <button
            type="submit"
            disabled={uploading || !file}
            className="btn btn-primary btn--full"
          >
            {uploading ? 'Загрузка...' : '📤 Загрузить документ'}
          </button>
        </form>

        {/* Ссылки */}
        <div className="page-auth__footer">
          <p className="page-auth__text">
            <Link to="/" className="link-muted">
              ← Вернуться к списку документов
            </Link>
          </p>
          <p className="page-auth__text" style={{ fontSize: '0.8125rem', marginTop: '0.5rem' }}>
            Поддерживаемые форматы: <strong>PDF, DOCX, TXT</strong>
          </p>
        </div>
      </div>
    </div>
  );
}