import { useState } from 'react'
import UploadGame from './features/games/UploadGame'
import GamesList from './features/games/GamesList'
import './App.css'

function App() {
  const [activeTab, setActiveTab] = useState<'upload' | 'list'>('upload')

  return (
    <div className="app">
      <header className="header">
        <h1>Mini Admin - Game Management</h1>
        <p>Upload and manage game files to S3</p>
      </header>

      <div className="tabs">
        <button
          className={activeTab === 'upload' ? 'tab active' : 'tab'}
          onClick={() => setActiveTab('upload')}
        >
          Upload Game
        </button>
        <button
          className={activeTab === 'list' ? 'tab active' : 'tab'}
          onClick={() => setActiveTab('list')}
        >
          Games List
        </button>
      </div>

      <div className="content">
        {activeTab === 'upload' ? <UploadGame /> : <GamesList />}
      </div>
    </div>
  )
}

export default App








