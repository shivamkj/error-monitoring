import { Link, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { listProjects } from '../../api/projects'

export default function Sidebar() {
  const location = useLocation()
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: listProjects,
  })

  return (
    <aside className="w-64 bg-gray-900 text-white flex flex-col">
      <div className="p-4 border-b border-gray-700">
        <Link to="/projects" className="text-xl font-bold tracking-tight">
          ErrorMonitor
        </Link>
      </div>
      <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
        <Link
          to="/projects"
          className={`block px-3 py-2 rounded text-sm ${
            location.pathname === '/projects'
              ? 'bg-gray-700 text-white'
              : 'text-gray-300 hover:bg-gray-800'
          }`}
        >
          All Projects
        </Link>
        {projects?.map((project) => (
          <Link
            key={project.id}
            to={`/projects/${project.id}/issues`}
            className={`block px-3 py-2 rounded text-sm truncate ${
              location.pathname.includes(project.id)
                ? 'bg-gray-700 text-white'
                : 'text-gray-300 hover:bg-gray-800'
            }`}
          >
            {project.name}
          </Link>
        ))}
      </nav>
    </aside>
  )
}
