import { apiFetch } from './client'

export interface ErrorEvent {
  id: string
  event_id: string
  issue_id: string
  project_id: string
  timestamp: string
  level: string
  platform: string
  ip_address: string
  user_data: Record<string, string> | null
  request_data: { url?: string; method?: string; headers?: Record<string, string> } | null
  breadcrumbs: Array<{
    timestamp: string
    level: string
    message: string
    category: string
    type: string
    data: Record<string, unknown>
  }> | null
  contexts: Record<string, Record<string, unknown>> | null
  tags: Record<string, string> | null
  exception: {
    values: Array<{
      type: string
      value: string
      stacktrace?: {
        frames: Array<{
          filename: string
          function: string
          lineno: number
          colno: number
          abs_path: string
          in_app?: boolean
        }>
      }
    }>
  } | null
  message: string
  environment: string
  release_tag: string
  server_name: string
  raw_payload: unknown
  created_at: string
}

interface EventsResponse {
  events: ErrorEvent[]
  total: number
  page: number
}

export async function listIssueEvents(
  issueId: string,
  params: { page?: number; per_page?: number } = {}
): Promise<EventsResponse> {
  const query = new URLSearchParams()
  if (params.page) query.set('page', String(params.page))
  if (params.per_page) query.set('per_page', String(params.per_page))
  const qs = query.toString()
  return apiFetch(`/issues/${issueId}/events${qs ? '?' + qs : ''}`)
}

export async function getEvent(id: string): Promise<ErrorEvent> {
  return apiFetch(`/events/${id}`)
}
