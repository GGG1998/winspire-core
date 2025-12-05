import { useState, useEffect } from 'react'
import { gamesApi, Game, UpdateGameRequest } from '../../api/games'
import './GamesList.css'

function GamesList() {
  const [games, setGames] = useState<Game[]>([])
  const [loading, setLoading] = useState(true)
  const [editingGame, setEditingGame] = useState<Game | null>(null)
  const [editForm, setEditForm] = useState<UpdateGameRequest>({})
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    loadGames()
  }, [])

  const loadGames = async () => {
    setLoading(true)
    try {
      const data = await gamesApi.getAllGames()
      setGames(data)
    } catch (error) {
      console.error('Failed to load games:', error)
      setMessage({ type: 'error', text: 'Failed to load games' })
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (game: Game) => {
    setEditingGame(game)
    setEditForm({
      slug: game.slug,
      name: game.name,
      description: game.description,
      logoUrl: game.logoUrl,
      version: game.version,
      versioningEnabled: game.versioningEnabled,
      isActive: game.isActive,
    })
  }

  const handleCancelEdit = () => {
    setEditingGame(null)
    setEditForm({})
  }

  const handleSaveEdit = async () => {
    if (!editingGame) return

    try {
      await gamesApi.updateGame(editingGame.id, editForm)
      setMessage({ type: 'success', text: 'Game updated successfully!' })
      setEditingGame(null)
      setEditForm({})
      loadGames()
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.error || 'Failed to update game',
      })
    }
  }

  const handleDelete = async (game: Game) => {
    if (!confirm(`Are you sure you want to delete "${game.name}"?`)) {
      return
    }

    try {
      await gamesApi.deleteGame(game.id)
      setMessage({ type: 'success', text: 'Game deleted successfully!' })
      loadGames()
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.error || 'Failed to delete game',
      })
    }
  }

  const handleGetUrl = async (game: Game) => {
    try {
      const { publicUrl } = await gamesApi.getGameUrl(game.id)
      navigator.clipboard.writeText(publicUrl)
      setMessage({ type: 'success', text: 'URL copied to clipboard!' })
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to get URL' })
    }
  }

  if (loading) {
    return (
      <div className="games-list">
        <h2>Games List</h2>
        <div className="loading">Loading games...</div>
      </div>
    )
  }

  return (
    <div className="games-list">
      <div className="header-row">
        <h2>Games List</h2>
        <button className="refresh-button" onClick={loadGames}>
          Refresh
        </button>
      </div>

      {message && (
        <div className={`message ${message.type}`}>
          {message.text}
          <button className="close-message" onClick={() => setMessage(null)}>
            ✕
          </button>
        </div>
      )}

      {games.length === 0 ? (
        <div className="empty-state">
          <p>No games found. Upload your first game!</p>
        </div>
      ) : (
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Slug</th>
                <th>Version</th>
                <th>Status</th>
                <th>Versioning</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {games.map((game) => (
                <tr key={game.id}>
                  <td>
                    <div className="game-name">
                      {game.name}
                      {game.description && (
                        <span className="game-description">{game.description}</span>
                      )}
                    </div>
                  </td>
                  <td>
                    <code>{game.slug}</code>
                  </td>
                  <td>{game.version}</td>
                  <td>
                    <span className={`status-badge ${game.isActive ? 'active' : 'inactive'}`}>
                      {game.isActive ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td>
                    <span className={`version-badge ${game.versioningEnabled ? 'enabled' : 'disabled'}`}>
                      {game.versioningEnabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </td>
                  <td>{new Date(game.createdAt).toLocaleDateString()}</td>
                  <td>
                    <div className="actions">
                      <button
                        className="action-button edit"
                        onClick={() => handleEdit(game)}
                        title="Edit"
                      >
                        ✏️
                      </button>
                      <button
                        className="action-button copy"
                        onClick={() => handleGetUrl(game)}
                        title="Copy URL"
                      >
                        🔗
                      </button>
                      <button
                        className="action-button delete"
                        onClick={() => handleDelete(game)}
                        title="Delete"
                      >
                        🗑️
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Edit Modal */}
      {editingGame && (
        <div className="modal-overlay" onClick={handleCancelEdit}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Edit Game: {editingGame.name}</h3>
              <button className="close-button" onClick={handleCancelEdit}>
                ✕
              </button>
            </div>

            <div className="modal-body">
              <div className="form-group">
                <label htmlFor="edit-name">Name</label>
                <input
                  type="text"
                  id="edit-name"
                  value={editForm.name || ''}
                  onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label htmlFor="edit-slug">Slug</label>
                <input
                  type="text"
                  id="edit-slug"
                  value={editForm.slug || ''}
                  onChange={(e) => setEditForm({ ...editForm, slug: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label htmlFor="edit-version">Version</label>
                <input
                  type="text"
                  id="edit-version"
                  value={editForm.version || ''}
                  onChange={(e) => setEditForm({ ...editForm, version: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label htmlFor="edit-description">Description</label>
                <textarea
                  id="edit-description"
                  value={editForm.description || ''}
                  onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                  rows={3}
                />
              </div>

              <div className="form-group">
                <label htmlFor="edit-logoUrl">Logo URL</label>
                <input
                  type="url"
                  id="edit-logoUrl"
                  value={editForm.logoUrl || ''}
                  onChange={(e) => setEditForm({ ...editForm, logoUrl: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={editForm.versioningEnabled || false}
                    onChange={(e) =>
                      setEditForm({ ...editForm, versioningEnabled: e.target.checked })
                    }
                  />
                  Enable S3 Versioning
                </label>
              </div>

              <div className="form-group">
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={editForm.isActive !== false}
                    onChange={(e) => setEditForm({ ...editForm, isActive: e.target.checked })}
                  />
                  Active
                </label>
              </div>
            </div>

            <div className="modal-footer">
              <button className="cancel-button" onClick={handleCancelEdit}>
                Cancel
              </button>
              <button className="save-button" onClick={handleSaveEdit}>
                Save Changes
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default GamesList

