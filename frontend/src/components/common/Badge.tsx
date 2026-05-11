const levelColors: Record<string, string> = {
  fatal: 'bg-red-100 text-red-800',
  error: 'bg-red-50 text-red-700',
  warning: 'bg-yellow-50 text-yellow-700',
  info: 'bg-blue-50 text-blue-700',
  debug: 'bg-gray-100 text-gray-700',
}

const statusColors: Record<string, string> = {
  unresolved: 'bg-orange-50 text-orange-700',
  resolved: 'bg-green-50 text-green-700',
  reappeared: 'bg-red-50 text-red-700',
  ignored: 'bg-gray-100 text-gray-600',
}

export function LevelBadge({ level }: { level: string }) {
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${levelColors[level] || levelColors.error}`}>
      {level}
    </span>
  )
}

export function StatusBadge({ status }: { status: string }) {
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${statusColors[status] || statusColors.unresolved}`}>
      {status}
    </span>
  )
}
