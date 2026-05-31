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
      setError('Выберите файл');
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
      setError(err.message || 'Ошибка загрузки');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div style={{ maxWidth: 600, margin: '2rem auto', padding: '1.5rem' }}>
      <header style={{ marginBottom: '1rem' }}>
        <Link to="/"><button>← Назад</button></Link>
        <h2 style={{ marginTop: '1rem' }}>➕ Загрузка документа</h2>
      </header>

      {error && (
        <p style={{ color: 'red', background: '#fee', padding: '0.5rem', borderRadius: '4px', marginBottom: '1rem' }}>
          {error}
        </p>
      )}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '1rem' }}>
          <label>Название (опционально)<br />
            <input 
              type="text" 
              value={title}
              onChange={e => setTitle(e.target.value)}
              placeholder="По умолчанию — имя файла"
              style={{ width: '100%', padding: '0.5rem', boxSizing: 'border-box' }}
            />
          </label>
        </div>

        <div style={{ marginBottom: '1rem' }}>
          <label>Файл *<br />
            <input 
              type="file" 
              onChange={e => setFile(e.target.files[0])}
              required
              style={{ width: '100%', padding: '0.5rem' }}
            />
          </label>
        </div>

        <div style={{ marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <input 
            type="checkbox" 
            id="isPublic"
            checked={isPublic}
            onChange={e => setIsPublic(e.target.checked)}
          />
          <label htmlFor="isPublic">🌐 Публичный документ (виден всем)</label>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button type="submit" disabled={uploading || !file} style={{ padding: '0.5rem 1rem' }}>
            {uploading ? 'Загрузка...' : '📤 Загрузить'}
          </button>
          <Link to="/"><button type="button">Отмена</button></Link>
        </div>
      </form>
    </div>
  );
}