/**
 * Human-friendly parser for standard 5-part cron expressions.
 */
export function describeCron(cron: string): string {
  if (!cron) return ''
  const trimmed = cron.trim()
  if (trimmed === '* * * * *') return 'Every minute'
  if (trimmed === '*/5 * * * *') return 'Every 5 minutes'
  if (trimmed === '*/10 * * * *') return 'Every 10 minutes'
  if (trimmed === '*/15 * * * *') return 'Every 15 minutes'
  if (trimmed === '*/30 * * * *') return 'Every 30 minutes'
  if (trimmed === '0 * * * *' || trimmed === '@hourly') return 'Every hour at minute 0'
  if (trimmed === '0 0 * * *' || trimmed === '@daily') return 'Every day at 00:00 (Midnight)'
  if (trimmed === '5 0 * * *') return 'Every day at 00:05'
  if (trimmed === '0 0 * * 0' || trimmed === '@weekly') return 'Every Sunday at 00:00'
  if (trimmed === '0 0 1 * *' || trimmed === '@monthly') return 'Day 1 of every month at 00:00'

  const parts = trimmed.split(/\s+/)
  if (parts.length === 5) {
    const [min, hour, dom, mon, dow] = parts
    if (dom === '*' && mon === '*' && dow === '*') {
      if (min.startsWith('*/')) return `Every ${min.slice(2)} minutes`
      if (hour !== '*') return `Daily at ${hour.padStart(2, '0')}:${min.padStart(2, '0')}`
    }
    if (mon === '*' && dow === '*' && dom !== '*') {
      return `Day ${dom} of month at ${hour.padStart(2, '0')}:${min.padStart(2, '0')}`
    }
    if (dom === '*' && mon === '*' && dow !== '*') {
      const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
      const dayName = days[Number(dow)] || `day ${dow}`
      return `Weekly on ${dayName} at ${hour.padStart(2, '0')}:${min.padStart(2, '0')}`
    }
  }

  return 'Custom cron expression'
}

export const CRON_PRESETS = [
  { label: 'Every 5 minutes', value: '*/5 * * * *' },
  { label: 'Every 15 minutes', value: '*/15 * * * *' },
  { label: 'Every hour (:00)', value: '0 * * * *' },
  { label: 'Every day at midnight (00:00)', value: '0 0 * * *' },
  { label: 'Every week on Sunday (00:00)', value: '0 0 * * 0' },
  { label: 'Every 1st of month (00:00)', value: '0 0 1 * *' },
]
