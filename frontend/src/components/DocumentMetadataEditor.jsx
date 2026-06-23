// frontend/src/components/DocumentMetadataEditor.jsx
import { useState, useEffect } from 'react';
import { documentsAPI } from '../services/api';

export default function DocumentMetadataEditor({ documentId, initialTitle, initialIsPublic, onUpdate, isOwner }) {
  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState(initialTitle);
  const [isPublic, setIsPublic] = useState(initialIsPublic);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  // Сброс при изменении пропсов
  useEffect(() => {
    setTitle(initialTitle);
    setIsPublic(initialIsPublic);
  }, [initialTitle, initialIsPublic]);

  const handleSave = async () => {
    if (!title.trim()) {
      setError('Название не может быть пустым');
      return;
    }
    
    setSaving(true);
    setError('');
    
    try {
      await documentsAPI.updateMetadata(documentId, {
        title: title.trim(),
        is_public: isPublic,
      });
      
      // Успех: коллбек + выход из режима редактирования
      onUpdate?.({ title: title.trim(), isPublic });
      setEditing(false);
      
    } catch (err) {
      console.error('Update metadata error:', err);
      setError(err.message || 'Не удалось сохранить изменения');
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    setTitle(initialTitle);
    setIsPublic(initialIsPublic);
    setError('');
    setEditing(false);
  };

  // Если не владелец — показываем только чтение
  if (!isOwner) {
    return (
      <div className="metadata-readonly">
        <h3 className="card__title">Настройки</h3>
        <dl className="meta-grid">
          <div className="meta-grid__item">
            <dt className="meta-grid__label">Название</dt>
            <dd className="meta-grid__value">{initialTitle}</dd>
          </div>
          <div className="meta-grid__item">
            <dt className="meta-grid__label">Видимость</dt>
            <dd className="meta-grid__value">
              {initialIsPublic ? (
                <span className="badge badge-success">Публичный</span>
              ) : (
                <span className="badge badge-muted">Приватный</span>
              )}
            </dd>
          </div>
        </dl>
      </div>
    );
  }

  // Режим редактирования
  if (editing) {
    return (
      <div className="card card--editor">
        <h3 className="card__title">Редактировать документ</h3>
        
        {error && (
          <div className="page-auth__error" role="alert">
            {error}
          </div>
        )}
        
        <div className="form-group">
          <label htmlFor="edit-title" className="form-label">Название документа</label>
          <input
            id="edit-title"
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="input"
            placeholder="Введите название"
            disabled={saving}
            autoFocus
          />
        </div>
        
        <div className="form-group">
          <label className="form-label">
            <input
              type="checkbox"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
              disabled={saving}
              style={{ marginRight: '0.5rem' }}
            />
            Сделать документ публичным (виден всем)
          </label>
        </div>
        
        <div className="modal-form__actions">
          <button
            type="button"
            onClick={handleCancel}
            className="btn btn-outline"
            disabled={saving}
          >
            Отмена
          </button>
          <button
            type="button"
            onClick={handleSave}
            className="btn btn-primary"
            disabled={saving || !title.trim()}
          >
            {saving ? 'Сохранение...' : 'Сохранить'}
          </button>
        </div>
      </div>
    );
  }

  // Режим просмотра (для владельца — с кнопкой редактирования)
  return (
    <div className="card card--info">
      <div className="card__header">
        <h3 className="card__title">Настройки</h3>
        <button 
          onClick={() => setEditing(true)}
          className="btn btn-ghost btn-sm"
          aria-label="Редактировать метаданные"
        >
          Изменить
        </button>
      </div>
      
      <dl className="meta-grid">
        <div className="meta-grid__item">
          <dt className="meta-grid__label">Название</dt>
          <dd className="meta-grid__value">{initialTitle}</dd>
        </div>
        <div className="meta-grid__item">
          <dt className="meta-grid__label">Видимость</dt>
          <dd className="meta-grid__value">
            {initialIsPublic ? (
              <span className="badge badge-success">Публичный</span>
            ) : (
              <span className="badge badge-muted">Приватный</span>
            )}
          </dd>
        </div>
      </dl>
    </div>
  );
}