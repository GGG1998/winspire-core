import { useState, useEffect, FormEvent, DragEvent } from 'react'
import { gamesApi, Game, CreateGameRequest } from '../../api/games'
import './UploadGame.css'

function UploadGame() {
  const [mode, setMode] = useState<'new' | 'existing'>('new')
  const [games, setGames] = useState<Game[]>([])
  const [selectedGameId, setSelectedGameId] = useState<string>('')
  const [formData, setFormData] = useState({
    slug: '',
    name: '',
    description: '',
    logoUrl: '',
    version: '1.0.0',
  })
  const [files, setFiles] = useState<File[]>([])
  const [isDragging, setIsDragging] = useState(false)
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  // Load existing games when mode is 'existing'
  useEffect(() => {
    if (mode === 'existing') {
      loadGames()
    }
  }, [mode])

  const loadGames = async () => {
    try {
      const data = await gamesApi.getAllGames()
      setGames(data.filter((g) => g.isActive))
    } catch (error) {
      console.error('Failed to load games:', error)
    }
  }

  const handleDragOver = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragging(true)
  }

  const handleDragLeave = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragging(false)
  }

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragging(false)

    const droppedFiles = Array.from(e.dataTransfer.files)
    setFiles((prev) => [...prev, ...droppedFiles])
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const selectedFiles = Array.from(e.target.files)
      setFiles((prev) => [...prev, ...selectedFiles])
    }
  }

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index))
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setMessage(null)

    if (files.length === 0) {
      setMessage({ type: 'error', text: 'Please select at least one file to upload.' })
      return
    }

    setLoading(true)

    try {
      if (mode === 'existing') {
        if (!selectedGameId) {
          setMessage({ type: 'error', text: 'Please select a game to update.' })
          return
        }

        const uploadResult = await gamesApi.uploadFiles(selectedGameId, files)
        setMessage({
          type: 'success',
          text: `Uploaded ${uploadResult.uploadedCount} file(s) for the selected game.`,
        })
        setFiles([])
        return
      }

      // New game workflow: upload files first, then create the game
      const uploadResult = await gamesApi.uploadNewGameFiles(
        formData.slug,
        formData.version,
        files
      )

      const createRequest: CreateGameRequest = {
        slug: formData.slug,
        name: formData.name,
        description: formData.description || undefined,
        logoUrl: formData.logoUrl || undefined,
        version: formData.version,
      }

      const newGame = await gamesApi.createGame(createRequest)

      setMessage({
        type: 'success',
        text: `Uploaded ${uploadResult.uploadedCount} file(s) and created "${newGame.name}".`,
      })
      setFiles([])
      setFormData({
        slug: '',
        name: '',
        description: '',
        logoUrl: '',
        version: '1.0.0',
      })
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.error || 'Failed to upload game',
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="upload-game">
      <h2>Upload Game</h2>

      <form onSubmit={handleSubmit}>
        {/* Mode Selection */}
        <div className="form-group">
          <label>Game Selection</label>
          <div className="radio-group">
            <label className="radio-label">
              <input
                type="radio"
                value="new"
                checked={mode === 'new'}
                onChange={() => setMode('new')}
              />
              New Game
            </label>
            <label className="radio-label">
              <input
                type="radio"
                value="existing"
                checked={mode === 'existing'}
                onChange={() => setMode('existing')}
              />
              Existing Game
            </label>
          </div>
        </div>

        {/* New Game Form */}
        {mode === 'new' && (
          <>
            <div className="form-group">
              <label htmlFor="slug">Game Slug *</label>
              <input
                type="text"
                id="slug"
                value={formData.slug}
                onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                placeholder="my-awesome-game"
                required
              />
            </div>

            <div className="form-group">
              <label htmlFor="name">Game Name *</label>
              <input
                type="text"
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="My Awesome Game"
                required
              />
            </div>

            <div className="form-group">
              <label htmlFor="version">Version *</label>
              <input
                type="text"
                id="version"
                value={formData.version}
                onChange={(e) => setFormData({ ...formData, version: e.target.value })}
                placeholder="1.0.0"
                required
              />
            </div>

            <div className="form-group">
              <label htmlFor="description">Description (Optional)</label>
              <textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Brief description of your game"
                rows={3}
              />
            </div>

            <div className="form-group">
              <label htmlFor="logoUrl">Logo URL (Optional)</label>
              <input
                type="url"
                id="logoUrl"
                value={formData.logoUrl}
                onChange={(e) => setFormData({ ...formData, logoUrl: e.target.value })}
                placeholder="https://example.com/logo.png"
              />
            </div>

          </>
        )}

        {/* Existing Game Dropdown */}
        {mode === 'existing' && (
          <div className="form-group">
            <label htmlFor="gameSelect">Select Game *</label>
            <select
              id="gameSelect"
              value={selectedGameId}
              onChange={(e) => setSelectedGameId(e.target.value)}
              required
            >
              <option value="">-- Choose a game --</option>
              {games.map((game) => (
                <option key={game.id} value={game.id}>
                  {game.name} ({game.slug})
                </option>
              ))}
            </select>
          </div>
        )}

        {/* File Drop Zone */}
        <div className="form-group">
          <label>Game Files</label>
          <div
            className={`drop-zone ${isDragging ? 'dragging' : ''}`}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
          >
            <div className="drop-zone-content">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="48"
                height="48"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
              <p>Drag & drop files here, or click to select</p>
              <input
                type="file"
                multiple
                onChange={handleFileSelect}
                style={{ display: 'none' }}
                id="fileInput"
              />
              <label htmlFor="fileInput" className="file-button">
                Choose Files
              </label>
            </div>
          </div>
        </div>

        {/* Selected Files List */}
        {files.length > 0 && (
          <div className="files-list">
            <h4>Selected Files ({files.length})</h4>
            <ul>
              {files.map((file, index) => (
                <li key={index}>
                  <span className="file-name">{file.name}</span>
                  <span className="file-size">
                    {(file.size / 1024).toFixed(2)} KB
                  </span>
                  <button
                    type="button"
                    onClick={() => removeFile(index)}
                    className="remove-button"
                  >
                    ✕
                  </button>
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Message */}
        {message && (
          <div className={`message ${message.type}`}>
            {message.text}
          </div>
        )}

        {/* Submit Button */}
        <button type="submit" className="submit-button" disabled={loading}>
          {loading ? 'Uploading...' : 'Upload'}
        </button>
      </form>
    </div>
  )
}

export default UploadGame
