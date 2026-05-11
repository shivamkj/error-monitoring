import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProject, deleteProject } from '../api/projects'
import { useState } from 'react'

export default function ProjectSettingsPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [copied, setCopied] = useState(false)

  const { data: project, isLoading } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => getProject(projectId!),
    enabled: !!projectId,
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteProject(projectId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      navigate('/projects')
    },
  })

  const copyDSN = () => {
    if (project?.dsn) {
      navigator.clipboard.writeText(project.dsn)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  if (isLoading) return <div className="text-gray-500">Loading...</div>
  if (!project) return <div className="text-gray-500">Project not found</div>

  return (
    <div className="max-w-2xl">
      <h1 className="text-2xl font-bold mb-6">{project.name} - Settings</h1>

      <div className="bg-white rounded-lg shadow border p-6 mb-6">
        <h2 className="text-lg font-semibold mb-3">DSN (Data Source Name)</h2>
        <p className="text-sm text-gray-600 mb-3">
          Use this DSN in your Sentry SDK configuration to send errors to this project.
        </p>
        <div className="flex items-center gap-2">
          <code className="flex-1 bg-gray-100 px-3 py-2 rounded text-sm font-mono break-all">
            {project.dsn}
          </code>
          <button
            onClick={copyDSN}
            className="px-3 py-2 bg-indigo-600 text-white rounded text-sm hover:bg-indigo-700"
          >
            {copied ? 'Copied!' : 'Copy'}
          </button>
        </div>

        <div className="mt-4 p-3 bg-gray-50 rounded text-sm">
          <p className="font-medium mb-2">Setup Example (@sentry/react):</p>
          <pre className="text-xs text-gray-700 overflow-x-auto">{`import * as Sentry from "@sentry/react";

Sentry.init({
  dsn: "${project.dsn}",
});`}</pre>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow border p-6">
        <h2 className="text-lg font-semibold mb-3 text-red-600">Danger Zone</h2>
        <p className="text-sm text-gray-600 mb-3">
          Deleting this project will permanently remove all issues and events.
        </p>
        <button
          onClick={() => {
            if (confirm('Are you sure you want to delete this project?')) {
              deleteMutation.mutate()
            }
          }}
          className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 text-sm"
        >
          Delete Project
        </button>
      </div>
    </div>
  )
}
