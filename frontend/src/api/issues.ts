import { apiFetch } from './client'

export interface Issue {
  id: string
  project_id: string
  fingerprint: string
  title: string
  culprit: string
  level: string
  platform: string
  status: string
  first_seen: string
  last_seen: string
  event_count: number
  browsers: Record<string, number>
  os_names: Record<string, number>
  devices: Record<string, number>
  urls: Record<string, number>
}

interface IssuesResponse {
  issues: Issue[]
  total: number
  page: number
}

export async function listIssues(
  projectId: string,
  params: { status?: string; sort?: string; page?: number; per_page?: number } = {}
): Promise<IssuesResponse> {
  const query = new URLSearchParams()
  if (params.status) query.set('status', params.status)
  if (params.sort) query.set('sort', params.sort)
  if (params.page) query.set('page', String(params.page))
  if (params.per_page) query.set('per_page', String(params.per_page))
  const qs = query.toString()
  return apiFetch(`/projects/${projectId}/issues${qs ? '?' + qs : ''}`)
}

export async function getIssue(id: string): Promise<Issue> {
  return apiFetch(`/issues/${id}`)
}

export async function updateIssueStatus(id: string, status: string): Promise<void> {
  return apiFetch(`/issues/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

export async function deleteIssue(id: string): Promise<void> {
  return apiFetch(`/issues/${id}`, { method: 'DELETE' })
}
