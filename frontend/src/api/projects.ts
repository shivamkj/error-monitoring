import { apiFetch } from './client'

export interface Project {
  id: string
  name: string
  slug: string
  platform: string
  public_key: string
  user_id: string
  dsn: string
  created_at: string
  updated_at: string
}

export async function listProjects(): Promise<Project[]> {
  return apiFetch('/projects')
}

export async function createProject(name: string, platform: string): Promise<Project> {
  return apiFetch('/projects', {
    method: 'POST',
    body: JSON.stringify({ name, platform }),
  })
}

export async function getProject(id: string): Promise<Project> {
  return apiFetch(`/projects/${id}`)
}

export async function deleteProject(id: string): Promise<void> {
  return apiFetch(`/projects/${id}`, { method: 'DELETE' })
}
