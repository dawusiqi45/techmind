import styles from './StatCard.module.css'

interface Props {
  title: string
  value: string | number
  status?: 'normal' | 'warning' | 'error'
  suffix?: string
}

export default function StatCard({ title, value, status = 'normal', suffix }: Props) {
  const colorMap = { normal: '#52c41a', warning: '#faad14', error: '#ff4d4f' }
  return (
    <div className={styles.card}>
      <div className={styles.title}>{title}</div>
      <div className={styles.value} style={{ color: colorMap[status] }}>
        {value}{suffix && <span className={styles.suffix}>{suffix}</span>}
      </div>
    </div>
  )
}
