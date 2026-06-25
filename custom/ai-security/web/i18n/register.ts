// 将 ai-security 模块的扁平 i18n 键注册进主应用 i18next 的 translation 命名空间。
// 主应用 config.ts 已设置 keySeparator:false，因此这里的带点扁平键（如 aiSecurity.name）
// 会按字面整体匹配，不会被当作嵌套路径。
import i18n from '@/i18n/config'
import resources from './ai-security.json'

let registered = false

export function registerAISecurityI18n() {
  if (registered) return
  registered = true
  const bundles = resources as Record<string, Record<string, string>>
  for (const lng of Object.keys(bundles)) {
    // deep=true 合并、overwrite=true 覆盖，确保与主 bundle 共存且以模块文案为准
    i18n.addResourceBundle(lng, 'translation', bundles[lng], true, true)
  }
}

registerAISecurityI18n()
