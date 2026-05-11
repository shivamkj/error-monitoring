export default function TimeAgo({ date }: { date: string }) {
  const seconds = Math.floor((Date.now() - new Date(date).getTime()) / 1000)

  if (seconds < 60) return <span>{seconds}s ago</span>
  if (seconds < 3600) return <span>{Math.floor(seconds / 60)}m ago</span>
  if (seconds < 86400) return <span>{Math.floor(seconds / 3600)}h ago</span>
  if (seconds < 604800) return <span>{Math.floor(seconds / 86400)}d ago</span>
  return <span>{new Date(date).toLocaleDateString()}</span>
}
