/** 将后端 UTC/RFC3339 时间格式化为 Asia/Shanghai：yyyy-MM-dd HH:mm:ss */
export function formatShanghai(iso: string): string {
  if (!iso) { return '-' }
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) { return iso }
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hourCycle: 'h23',
  }).formatToParts(d)
  const get = (type: Intl.DateTimeFormatPartTypes) =>
    parts.find(p => p.type === type)?.value || '00'
  return `${get('year')}-${get('month')}-${get('day')} ${get('hour')}:${get('minute')}:${get('second')}`
}
