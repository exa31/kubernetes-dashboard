/**
 * Composable: useDarkMode
 * Handles dark mode toggle with localStorage persistence.
 */
import { ref, watchEffect } from 'vue'

const savedTheme = typeof localStorage !== 'undefined' ? localStorage.getItem('theme') : null
// Default to dark mode for Kubernetes dashboard if not explicitly set
const isDark = ref(savedTheme !== null ? savedTheme === 'dark' : true)

function applyTheme(dark: boolean) {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  if (dark) {
    root.classList.add('dark')
    root.style.colorScheme = 'dark'
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('theme', 'dark')
    }
  } else {
    root.classList.remove('dark')
    root.style.colorScheme = 'light'
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('theme', 'light')
    }
  }
}

// Apply immediately upon script evaluation
applyTheme(isDark.value)

export function useDarkMode() {
  watchEffect(() => {
    applyTheme(isDark.value)
  })

  const toggle = () => {
    isDark.value = !isDark.value
    applyTheme(isDark.value)
  }

  const setLight = () => {
    isDark.value = false
    applyTheme(false)
  }

  const setDark = () => {
    isDark.value = true
    applyTheme(true)
  }

  return { isDark, toggle, setLight, setDark }
}
