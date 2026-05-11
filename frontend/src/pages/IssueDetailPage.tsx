import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getIssue, updateIssueStatus, deleteIssue } from '../api/issues'
import { listIssueEvents, type ErrorEvent } from '../api/events'
import { LevelBadge, StatusBadge } from '../components/common/Badge'
import TimeAgo from '../components/common/TimeAgo'

export default function IssueDetailPage() {
  const { issueId } = useParams<{ issueId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<'stacktrace' | 'events' | 'context'>('stacktrace')
  const [selectedEvent, setSelectedEvent] = useState<ErrorEvent | null>(null)

  const { data: issue } = useQuery({
    queryKey: ['issue', issueId],
    queryFn: () => getIssue(issueId!),
    enabled: !!issueId,
  })

  const { data: eventsData } = useQuery({
    queryKey: ['issue-events', issueId],
    queryFn: () => listIssueEvents(issueId!),
    enabled: !!issueId,
  })

  const statusMutation = useMutation({
    mutationFn: (status: string) => updateIssueStatus(issueId!, status),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['issue', issueId] }),
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteIssue(issueId!),
    onSuccess: () => navigate(-1),
  })

  if (!issue) return <div className="text-gray-500">Loading...</div>

  const latestEvent = selectedEvent || eventsData?.events[0]

  return (
    <div>
      <div className="mb-6">
        <div className="flex items-center gap-2 mb-2">
          <LevelBadge level={issue.level} />
          <StatusBadge status={issue.status} />
        </div>
        <h1 className="text-xl font-bold text-gray-900 mb-1">{issue.title}</h1>
        {issue.culprit && <p className="text-sm text-gray-500">{issue.culprit}</p>}
        <div className="flex items-center gap-4 mt-3 text-sm text-gray-500">
          <span>Events: <strong>{issue.event_count}</strong></span>
          <span>First: <TimeAgo date={issue.first_seen} /></span>
          <span>Last: <TimeAgo date={issue.last_seen} /></span>
        </div>
      </div>

      <div className="flex gap-2 mb-6">
        {issue.status !== 'resolved' && (
          <button
            onClick={() => statusMutation.mutate('resolved')}
            className="px-3 py-1.5 bg-green-600 text-white rounded text-sm hover:bg-green-700"
          >
            Resolve
          </button>
        )}
        {issue.status !== 'ignored' && (
          <button
            onClick={() => statusMutation.mutate('ignored')}
            className="px-3 py-1.5 bg-gray-600 text-white rounded text-sm hover:bg-gray-700"
          >
            Ignore
          </button>
        )}
        {issue.status !== 'unresolved' && (
          <button
            onClick={() => statusMutation.mutate('unresolved')}
            className="px-3 py-1.5 border border-gray-300 rounded text-sm hover:bg-gray-50"
          >
            Unresolve
          </button>
        )}
        <button
          onClick={() => { if (confirm('Delete this issue?')) deleteMutation.mutate() }}
          className="px-3 py-1.5 bg-red-600 text-white rounded text-sm hover:bg-red-700 ml-auto"
        >
          Delete
        </button>
      </div>

      <div className="flex gap-1 border-b mb-4">
        {(['stacktrace', 'events', 'context'] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px capitalize ${
              activeTab === tab
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      {activeTab === 'stacktrace' && latestEvent && (
        <StackTraceView event={latestEvent} />
      )}

      {activeTab === 'events' && (
        <EventsList
          events={eventsData?.events || []}
          onSelect={setSelectedEvent}
          selectedId={selectedEvent?.id}
        />
      )}

      {activeTab === 'context' && (
        <ContextView issue={issue} event={latestEvent} />
      )}
    </div>
  )
}

function StackTraceView({ event }: { event: ErrorEvent }) {
  if (!event.exception?.values?.length) {
    return <p className="text-gray-500 text-sm">No stack trace available</p>
  }

  return (
    <div className="space-y-4">
      {event.exception.values.map((exc, i) => (
        <div key={i} className="bg-white rounded-lg shadow border overflow-hidden">
          <div className="px-4 py-3 bg-red-50 border-b">
            <span className="font-semibold text-red-800">{exc.type}</span>
            {exc.value && <span className="text-red-600 ml-2">{exc.value}</span>}
          </div>
          {exc.stacktrace?.frames && (
            <div className="divide-y">
              {[...exc.stacktrace.frames].reverse().map((frame, j) => (
                <div
                  key={j}
                  className={`px-4 py-2 text-sm ${frame.in_app ? 'bg-white' : 'bg-gray-50'}`}
                >
                  <div className="flex items-center gap-2">
                    {frame.in_app && (
                      <span className="w-2 h-2 rounded-full bg-indigo-500 flex-shrink-0" />
                    )}
                    <span className="font-mono text-gray-700 truncate">
                      {frame.filename || frame.abs_path}
                    </span>
                    {frame.lineno > 0 && (
                      <span className="text-gray-400">:{frame.lineno}</span>
                    )}
                  </div>
                  {frame.function && (
                    <p className="text-gray-500 font-mono ml-4 text-xs">{frame.function}</p>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      ))}

      {event.breadcrumbs && event.breadcrumbs.length > 0 && (
        <div className="bg-white rounded-lg shadow border overflow-hidden mt-4">
          <div className="px-4 py-3 bg-gray-50 border-b font-medium text-sm">Breadcrumbs</div>
          <div className="divide-y max-h-96 overflow-y-auto">
            {event.breadcrumbs.map((bc, i) => (
              <div key={i} className="px-4 py-2 text-sm">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-gray-400 font-mono w-20 flex-shrink-0">
                    {bc.category || bc.type}
                  </span>
                  <span className="text-gray-700 truncate">{bc.message}</span>
                  {bc.level && <LevelBadge level={bc.level} />}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function EventsList({ events, onSelect, selectedId }: {
  events: ErrorEvent[]
  onSelect: (e: ErrorEvent) => void
  selectedId?: string
}) {
  return (
    <div className="bg-white rounded-lg shadow border overflow-hidden">
      <div className="divide-y">
        {events.map((event) => (
          <button
            key={event.id}
            onClick={() => onSelect(event)}
            className={`w-full text-left px-4 py-3 hover:bg-gray-50 ${
              selectedId === event.id ? 'bg-indigo-50' : ''
            }`}
          >
            <div className="flex items-center justify-between">
              <span className="text-sm font-mono text-gray-600">{event.event_id.slice(0, 12)}</span>
              <TimeAgo date={event.timestamp} />
            </div>
            <div className="flex items-center gap-2 mt-1 text-xs text-gray-500">
              {event.ip_address && <span>{event.ip_address}</span>}
              {event.environment && <span className="bg-gray-100 px-1.5 py-0.5 rounded">{event.environment}</span>}
              {event.release_tag && <span>v{event.release_tag}</span>}
            </div>
          </button>
        ))}
      </div>
      {events.length === 0 && (
        <p className="p-4 text-gray-500 text-sm text-center">No events</p>
      )}
    </div>
  )
}

function ContextView({ issue, event }: { issue: ReturnType<typeof getIssue> extends Promise<infer T> ? T : never; event?: ErrorEvent | null }) {
  return (
    <div className="space-y-4">
      <div className="bg-white rounded-lg shadow border p-4">
        <h3 className="font-semibold text-sm mb-3">Aggregate Stats</h3>
        <div className="grid grid-cols-2 gap-4">
          <AggregateSection title="Browsers" data={issue.browsers} />
          <AggregateSection title="Operating Systems" data={issue.os_names} />
          <AggregateSection title="Devices" data={issue.devices} />
          <AggregateSection title="URLs" data={issue.urls} />
        </div>
      </div>

      {event?.user_data && Object.keys(event.user_data).length > 0 && (
        <div className="bg-white rounded-lg shadow border p-4">
          <h3 className="font-semibold text-sm mb-2">User</h3>
          <dl className="text-sm space-y-1">
            {Object.entries(event.user_data).map(([k, v]) => (
              <div key={k} className="flex gap-2">
                <dt className="text-gray-500 w-24">{k}:</dt>
                <dd className="text-gray-700">{v}</dd>
              </div>
            ))}
          </dl>
        </div>
      )}

      {event?.request_data && (
        <div className="bg-white rounded-lg shadow border p-4">
          <h3 className="font-semibold text-sm mb-2">Request</h3>
          <dl className="text-sm space-y-1">
            {event.request_data.method && (
              <div className="flex gap-2">
                <dt className="text-gray-500 w-24">Method:</dt>
                <dd className="text-gray-700">{event.request_data.method}</dd>
              </div>
            )}
            {event.request_data.url && (
              <div className="flex gap-2">
                <dt className="text-gray-500 w-24">URL:</dt>
                <dd className="text-gray-700 break-all">{event.request_data.url}</dd>
              </div>
            )}
          </dl>
        </div>
      )}

      {event?.tags && Object.keys(event.tags).length > 0 && (
        <div className="bg-white rounded-lg shadow border p-4">
          <h3 className="font-semibold text-sm mb-2">Tags</h3>
          <div className="flex flex-wrap gap-2">
            {Object.entries(event.tags).map(([k, v]) => (
              <span key={k} className="bg-gray-100 px-2 py-1 rounded text-xs">
                <span className="text-gray-500">{k}:</span> {v}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function AggregateSection({ title, data }: { title: string; data: Record<string, number> }) {
  const entries = Object.entries(data || {}).sort((a, b) => b[1] - a[1])
  if (entries.length === 0) return null

  const total = entries.reduce((sum, [, count]) => sum + count, 0)

  return (
    <div>
      <h4 className="text-xs text-gray-500 font-medium mb-1">{title}</h4>
      <div className="space-y-1">
        {entries.slice(0, 5).map(([name, count]) => (
          <div key={name} className="flex items-center gap-2">
            <div className="flex-1 bg-gray-100 rounded-full h-2 overflow-hidden">
              <div
                className="h-full bg-indigo-500 rounded-full"
                style={{ width: `${(count / total) * 100}%` }}
              />
            </div>
            <span className="text-xs text-gray-600 w-24 truncate">{name}</span>
            <span className="text-xs text-gray-400 w-8 text-right">{count}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
