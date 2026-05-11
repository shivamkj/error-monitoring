import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listProjects, createProject } from '../api/projects'

export default function ProjectsPage() {
  const queryClient = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [platform, setPlatform] = useState('javascript')

  const { data: projects, isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: listProjects,
  })

  const createMutation = useMutation({
    mutationFn: () => createProject(name, platform),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      setShowCreate(false)
      setName('')
    },
  })

  if (isLoading) return <div className="text-gray-500">Loading...</div>

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Projects</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700 text-sm font-medium"
        >
          New Project
        </button>
      </div>

      {showCreate && (
        <div className="mb-6 p-4 bg-white rounded-lg shadow border">
          <form onSubmit={(e) => { e.preventDefault(); createMutation.mutate() }} className="flex gap-3 items-end">
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                placeholder="My App"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Platform</label>
              <select
                value={platform}
                onChange={(e) => setPlatform(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-md"
              >
                <option value="javascript">JavaScript</option>
                <option value="react">React</option>
                <option value="node">Node.js</option>
                <option value="python">Python</option>
                <option value="flutter">Flutter</option>
                <option value="go">Go</option>
              </select>
            </div>
            <button type="submit" className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700 text-sm">
              Create
            </button>
            <button type="button" onClick={() => setShowCreate(false)} className="text-gray-500 px-4 py-2 text-sm">
              Cancel
            </button>
          </form>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {projects?.map((project) => (
          <Link
            key={project.id}
            to={`/projects/${project.id}/issues`}
            className="block p-4 bg-white rounded-lg shadow border hover:border-indigo-300 transition-colors"
          >
            <h2 className="font-semibold text-lg">{project.name}</h2>
            <p className="text-sm text-gray-500 mt-1">{project.platform}</p>
            <div className="mt-3 flex items-center justify-between">
              <span className="text-xs text-gray-400">
                Created {new Date(project.created_at).toLocaleDateString()}
              </span>
              <Link
                to={`/projects/${project.id}/settings`}
                onClick={(e) => e.stopPropagation()}
                className="text-xs text-indigo-600 hover:underline"
              >
                Settings
              </Link>
            </div>
          </Link>
        ))}
      </div>

      {projects?.length === 0 && (
        <p className="text-gray-500 text-center mt-8">No projects yet. Create one to get started.</p>
      )}
    </div>
  )
}
