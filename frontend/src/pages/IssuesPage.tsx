import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { listIssues } from '../api/issues'
import { LevelBadge, StatusBadge } from '../components/common/Badge'
import TimeAgo from '../components/common/TimeAgo'

const STATUS_TABS = [
  { value: '', label: 'All' },
  { value: 'unresolved', label: 'Unresolved' },
  { value: 'reappeared', label: 'Reappeared' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'ignored', label: 'Ignored' },
]

const SORT_OPTIONS = [
  { value: 'last_seen', label: 'Last Seen' },
  { value: 'first_seen', label: 'First Seen' },
  { value: 'event_count', label: 'Events' },
]

export default function IssuesPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const [status, setStatus] = useState('')
  const [sort, setSort] = useState('last_seen')
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['issues', projectId, status, sort, page],
    queryFn: () => listIssues(projectId!, { status: status || undefined, sort, page }),
    enabled: !!projectId,
  })

  if (isLoading) return <div className="text-gray-500">Loading...</div>

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Issues</h1>
        <select
          value={sort}
          onChange={(e) => setSort(e.target.value)}
          className="text-sm border border-gray-300 rounded px-2 py-1"
        >
          {SORT_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
      </div>

      <div className="flex gap-1 mb-4 border-b">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => { setStatus(tab.value); setPage(1) }}
            className={`px-3 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              status === tab.value
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="bg-white rounded-lg shadow border overflow-hidden">
        {data?.issues.length === 0 ? (
          <p className="p-8 text-center text-gray-500">No issues found</p>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b bg-gray-50 text-left text-xs text-gray-500 uppercase">
                <th className="px-4 py-3">Issue</th>
                <th className="px-4 py-3 w-20">Events</th>
                <th className="px-4 py-3 w-32">Last Seen</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {data?.issues.map((issue) => (
                <tr key={issue.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <Link to={`/issues/${issue.id}`} className="block">
                      <div className="flex items-center gap-2 mb-1">
                        <LevelBadge level={issue.level} />
                        <StatusBadge status={issue.status} />
                      </div>
                      <p className="font-medium text-sm text-gray-900 truncate max-w-lg">
                        {issue.title}
                      </p>
                      {issue.culprit && (
                        <p className="text-xs text-gray-500 truncate mt-0.5">{issue.culprit}</p>
                      )}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600 font-mono">
                    {issue.event_count.toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    <TimeAgo date={issue.last_seen} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {data && data.total > 25 && (
        <div className="flex justify-center gap-2 mt-4">
          <button
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page === 1}
            className="px-3 py-1 border rounded text-sm disabled:opacity-50"
          >
            Previous
          </button>
          <span className="px-3 py-1 text-sm text-gray-600">
            Page {page} of {Math.ceil(data.total / 25)}
          </span>
          <button
            onClick={() => setPage(page + 1)}
            disabled={page >= Math.ceil(data.total / 25)}
            className="px-3 py-1 border rounded text-sm disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
