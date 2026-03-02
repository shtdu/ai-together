export type AutomationCard = {
  id: number
  name: string
  apiUrl: string
  apiKey: string
  tint: string
  accent: string
  enabled: boolean
  // Priority grouping - Smaller number = higher priority (1-10, default 1)
  level?: number
  // Team ID from server (used when pushing to server)
  teamId?: number
  // 模型白名单：声明 provider 支持的模型（精确或通配符）
  supportedModels?: string[]
  // 模型映射：external model -> internal model
  modelMapping?: Record<string, string>
}

export const automationCardGroups: Record<'claude' | 'codex' | 'opencode', AutomationCard[]> = {
  claude: [
    {
      id: 100,
      name: '0011',
      apiUrl: 'https://0011.ai',
      apiKey: '',
      tint: 'rgba(10, 132, 255, 0.14)',
      accent: '#0aff5cff',
      enabled: false,
    },
    {
      id: 101,
      name: 'AICoding.sh',
      apiUrl: 'https://api.aicoding.sh',
      apiKey: '',
      tint: 'rgba(10, 132, 255, 0.14)',
      accent: '#0a84ff',
      enabled: false,
    },
    {
      id: 102,
      name: 'Kimi',
      apiUrl: 'https://api.moonshot.cn/anthropic',
      apiKey: '',
      tint: 'rgba(16, 185, 129, 0.16)',
      accent: '#10b981',
      enabled: false,
    },
    {
      id: 103,
      name: 'Deepseek',
      apiUrl: 'https://api.deepseek.com/anthropic',
      apiKey: '',
      tint: 'rgba(251, 146, 60, 0.18)',
      accent: '#f97316',
      enabled: false,
    },
  ],
  codex: [
    {
      id: 201,
      name: 'AICoding.sh',
      apiUrl: 'https://api.aicoding.sh',
      apiKey: '',
      tint: 'rgba(236, 72, 153, 0.16)',
      accent: '#ec4899',
      enabled: false,
    },
  ],
  opencode: [
    {
      id: 301,
      name: 'AICoding.sh',
      apiUrl: 'https://api.aicoding.sh',
      apiKey: '',
      tint: 'rgba(99, 102, 241, 0.16)',
      accent: '#6366f1',
      enabled: false,
    },
  ],
}

export function createAutomationCards(data: AutomationCard[] = []): AutomationCard[] {
  // First pass: assign levels based on array position for items without level
  return data
    .map((item, index) => ({
      ...item,
      // Default to array position + 1 if not set (preserves file order)
      level: item.level ?? (index + 1),
      // Ensure modelMapping and supportedModels are always initialized
      // to prevent data loss during serialization
      modelMapping: item.modelMapping ?? {},
      supportedModels: item.supportedModels ?? [],
    }))
    .sort((a, b) => {
      const levelA = a.level ?? 1
      const levelB = b.level ?? 1
      if (levelA !== levelB) {
        return levelA - levelB // Lower level = higher priority
      }
      // Preserve original order within same level (by id)
      return a.id - b.id
    })
}
