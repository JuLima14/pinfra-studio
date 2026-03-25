import { Routes, Route } from 'react-router-dom'
import { ProjectList } from '@/components/ProjectList'
import { Layout } from '@/components/Layout'

function App() {
  return (
    <Routes>
      <Route path="/" element={<ProjectList />} />
      <Route path="/projects/:id" element={<Layout />} />
    </Routes>
  )
}

export default App
