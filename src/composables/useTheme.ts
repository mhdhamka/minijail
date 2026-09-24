import { ref, watch, onMounted } from 'vue';

export type Theme = 'dark' | 'light';

const theme = ref<Theme>('dark');

export function useTheme() {
  function applyTheme(t: Theme) {
    theme.value = t;
    if (typeof document !== 'undefined') {
      const root = document.documentElement;
      if (t === 'light') {
        root.classList.add('light-theme');
        root.classList.remove('dark-theme');
        document.body.classList.add('light-mode');
      } else {
        root.classList.remove('light-theme');
        root.classList.add('dark-theme');
        document.body.classList.remove('light-mode');
      }
      try {
        localStorage.setItem('docker-simulator-theme', t);
      } catch (e) {
        // ignore in private browsing
      }
    }
  }

  function toggleTheme() {
    applyTheme(theme.value === 'dark' ? 'light' : 'dark');
  }

  onMounted(() => {
    try {
      const saved = localStorage.getItem('docker-simulator-theme') as Theme | null;
      if (saved === 'light' || saved === 'dark') {
        applyTheme(saved);
      } else {
        applyTheme('dark');
      }
    } catch (e) {
      applyTheme('dark');
    }
  });

  return {
    theme,
    toggleTheme,
    applyTheme
  };
}
